package collection_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/sync/errgroup"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
)

const (
	defaultCollectionSize = 10
	originalValue         = "old"
	changedValue          = "new"
	collectionName        = "test_items"
)

var (
	TestSecondaryIndex = &col.Index{Field: "Value"}
)

type TestError struct{}

func (te TestError) Is(other error) bool {
	_, ok := other.(TestError)
	return ok
}

func (te TestError) Error() string {
	return "TestError"
}

type ReadCallback func(context.Context) col.ReadOnlyCollection
type WriteCallback func(context.Context, func(col.ReadWriteCollection) error) error

func idRange(start int, end int) []string {
	result := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, makeID(i))
	}
	return result
}

func doWrite(id string) func(rw col.ReadWriteCollection) error {
	return func(rw col.ReadWriteCollection) error {
		testProto := makeProto(id)
		return rw.Put(testProto.ID, testProto)
	}
}

func doDelete(id string) func(rw col.ReadWriteCollection) error {
	return func(rw col.ReadWriteCollection) error {
		return rw.Delete(id)
	}
}

func doDeleteAll() func(rw col.ReadWriteCollection) error {
	return func(rw col.ReadWriteCollection) error {
		return rw.DeleteAll()
	}
}

// canceledContext is a helper function to provide a context that is already canceled
func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// populateCollection writes a standard set of rows to the collection for use by
// individual tests.
func populateCollection(rw col.ReadWriteCollection) error {
	for _, id := range idRange(0, defaultCollectionSize) {
		testProto := makeProto(id)
		testProto.Value = originalValue
		if err := rw.Create(testProto.ID, testProto); err != nil {
			return err
		}
	}
	return nil
}

// Helper function to turn an int ID into a string so we don't need to use string literals
func makeID(i int) string {
	return fmt.Sprintf("%d", i)
}

// Helper function to instantiate the proto for a row by ID
func makeProto(id string) *col.TestItem {
	return &col.TestItem{ID: id, Value: changedValue}
}

type RowDiff struct {
	Deleted []string
	Changed []string
	Created []string
}

// checkDefaultCollection validates that the contents of the collection match
// the expected set of rows based on a diff from the default collection
// populated by populateCollection.
func checkDefaultCollection(t *testing.T, ro col.ReadOnlyCollection, diff RowDiff) {
	expected := map[string]string{}
	for _, id := range idRange(0, defaultCollectionSize) {
		expected[id] = originalValue
	}
	if diff.Deleted != nil {
		for _, id := range diff.Deleted {
			_, ok := expected[id]
			require.True(t, ok, "test specified a deleted row that was not in the original set")
			delete(expected, id)
		}
	}
	if diff.Changed != nil {
		for _, id := range diff.Changed {
			_, ok := expected[id]
			require.True(t, ok, "test specified a changed row that was not in the original set")
			expected[id] = changedValue
		}
	}
	if diff.Created != nil {
		for _, id := range diff.Created {
			_, ok := expected[id]
			require.False(t, ok, "test specified an added row that was already in the original set")
			expected[id] = changedValue
		}
	}
	checkCollection(t, ro, expected)
}

func checkItem(t *testing.T, item *col.TestItem, id string, value string) error {
	if item.ID != id {
		return errors.Errorf("Item has incorrect ID, expected: %s, actual: %s", id, item.ID)
	}
	if item.Value != value {
		return errors.Errorf("Item has incorrect Value, expected: %s, actual: %s", value, item.Value)
	}
	return nil
}

func checkCollection(t *testing.T, ro col.ReadOnlyCollection, expected map[string]string) {
	testProto := &col.TestItem{}
	actual := map[string]string{}
	require.NoError(t, ro.List(testProto, col.DefaultOptions(), func() error {
		actual[testProto.ID] = testProto.Value
		return nil
	}))

	for k, v := range expected {
		other, ok := actual[k]
		require.True(t, ok, "row '%s' was expected to be present but was not found", k)
		require.Equal(t, v, other, "row '%s' had an unexpected value", k)
	}

	for k := range actual {
		_, ok := expected[k]
		require.True(t, ok, "row '%s' was present but was not expected", k)
	}
}

func collectionTests(
	parent *testing.T,
	newCollection func(context.Context, *testing.T) (ReadCallback, WriteCallback),
) {
	parent.Run("ReadOnly", func(suite *testing.T) {
		suite.Parallel()
		emptyReader, _ := newCollection(context.Background(), suite)
		emptyRead := emptyReader(context.Background())
		defaultReader, writer := newCollection(context.Background(), suite)
		defaultRead := defaultReader(context.Background())
		require.NoError(suite, writer(context.Background(), populateCollection))

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
				itemProto := &col.TestItem{}
				err := defaultRead.Get("baz", itemProto)
				require.True(t, errors.Is(err, col.ErrNotFound{}), "Incorrect error: %v", err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				itemProto := &col.TestItem{}
				require.NoError(t, defaultRead.Get("5", itemProto))
				require.Equal(t, "5", itemProto.ID)
			})
		})

		suite.Run("GetByIndex", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("List", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Empty", func(t *testing.T) {
				t.Parallel()
				testProto := &col.TestItem{}
				err := emptyRead.List(testProto, col.DefaultOptions(), func() error {
					return errors.Errorf("List callback should not have been called for an empty collection")
				})
				require.NoError(t, err)
			})

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				items := map[string]string{}
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func() error {
					items[testProto.ID] = testProto.Value
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.ID = ""
					testProto.Value = ""
					return nil
				})
				require.NoError(t, err)

				keys := []string{}
				for k, v := range items {
					keys = append(keys, k)
					require.Equal(t, originalValue, v)
				}
				expectedKeys := idRange(0, defaultCollectionSize)
				require.ElementsEqual(t, expectedKeys, keys)
			})

			subsuite.Run("Sort", func(subsuite *testing.T) {
				subsuite.Parallel()

				testSort := func(t *testing.T, ro col.ReadOnlyCollection, target col.SortTarget, expectedAsc []string) {
					t.Parallel()

					collect := func(order col.SortOrder) []string {
						keys := []string{}
						testProto := &col.TestItem{}
						err := ro.List(testProto, &col.Options{Target: target, Order: order}, func() error {
							keys = append(keys, testProto.ID)
							return nil
						})
						require.NoError(t, err)
						return keys
					}

					t.Run("SortAscend", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortAscend)
						require.Equal(t, expectedAsc, keys)
					})

					t.Run("SortDescend", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortDescend)
						reversed := []string{}
						for i := len(expectedAsc) - 1; i >= 0; i-- {
							reversed = append(reversed, expectedAsc[i])
						}
						require.Equal(t, reversed, keys)
					})

					t.Run("SortNone", func(t *testing.T) {
						t.Parallel()
						keys := collect(col.SortNone)
						require.ElementsEqual(t, expectedAsc, keys)
					})
				}

				subsuite.Run("CreatedAt", func(t *testing.T) {
					// The default collection was written in a single transaction, so all
					// the rows have the same created time - make our own
					createReader, writer := newCollection(context.Background(), t)
					createCol := createReader(context.Background())

					createKeys := []string{"0", "6", "7", "9", "3", "8", "4", "1", "2", "5"}
					for _, k := range createKeys {
						err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
							testProto := &col.TestItem{ID: k, Value: originalValue}
							return rw.Create(k, testProto)
						})
						require.NoError(t, err)
					}

					checkDefaultCollection(t, createCol, RowDiff{})
					testSort(t, createCol, col.SortByCreateRevision, createKeys)
				})

				subsuite.Run("UpdatedAt", func(t *testing.T) {
					// Create a new collection that we can modify to get a different ordering here
					reader, writer := newCollection(context.Background(), t)
					modRead := reader(context.Background())
					require.NoError(suite, writer(context.Background(), populateCollection))

					modKeys := []string{"5", "7", "2", "9", "1", "0", "8", "4", "3", "6"}
					for _, k := range modKeys {
						err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
							testProto := &col.TestItem{}
							if err := rw.Update(k, testProto, func() error {
								testProto.Value = changedValue
								return nil
							}); err != nil {
								return err
							}
							return nil
						})
						require.NoError(t, err)
					}
					testSort(t, modRead, col.SortByModRevision, modKeys)
				})

				subsuite.Run("Key", func(t *testing.T) {
					testSort(t, defaultRead, col.SortByKey, idRange(0, 10))
				})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				count := 0
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func() error {
					count++
					return &TestError{}
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				require.Equal(t, 1, count, "List callback was called multiple times despite erroring")
			})

			subsuite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				count := 0
				testProto := &col.TestItem{}
				err := defaultRead.List(testProto, col.DefaultOptions(), func() error {
					count++
					return errutil.ErrBreak
				})
				require.NoError(t, err)
				require.Equal(t, 1, count, "List callback was called multiple times despite aborting")
			})
		})

		suite.Run("Count", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				count, err := defaultRead.Count()
				require.NoError(t, err)
				require.Equal(t, int64(10), count)

				count, err = emptyRead.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})

		// 'Watch' functions with callbacks don't have a good hook to trigger
		// events. This is a bit racy, but we can sleep a bit to let the listen get
		// started, then trigger some writes.
		asyncWrite := func(t *testing.T, writer WriteCallback, cbs ...func(rw col.ReadWriteCollection) error) {
			eg, ctx := errgroup.WithContext(context.Background())

			eg.Go(func() error {
				time.Sleep(100 * time.Millisecond)
				for _, cb := range cbs {
					if err := writer(ctx, cb); err != nil {
						return err
					}
				}
				return nil
			})

			t.Cleanup(func() {
				require.NoError(t, eg.Wait())
			})
		}

		// Helper function for checking watch events
		collectEventsCallback := func(count int, out *[]*watch.Event) func(*watch.Event) error {
			return func(ev *watch.Event) error {
				fmt.Printf("got event: %v\n", ev)
				*out = append(*out, ev)
				if len(*out) == count {
					return errutil.ErrBreak
				}
				return nil
			}
		}

		collectEventsChannel := func(watcher watch.Watcher, count int, out *[]*watch.Event) {
			for len(*out) < count {
				*out = append(*out, <-watcher.Watch())
			}
		}

		type testEvent struct {
			Type  watch.EventType
			Key   string
			Value *col.TestItem
		}

		checkEvents := func(t *testing.T, actual []*watch.Event, expected []testEvent) {
			require.Equal(t, len(actual), len(expected), "Incorrect number of watch events")
			for i := 0; i < len(actual); i++ {
				require.Equal(t, expected[i].Type, actual[i].Type)
				if actual[i].Type != watch.EventError {
					require.Equal(t, expected[i].Key, string(actual[i].Key))
				}
				if actual[i].Type == watch.EventPut {
					actualProto := &col.TestItem{}
					require.NoError(t, proto.Unmarshal(actual[i].Value, actualProto))
					require.Equal(t, expected[i].Value, actualProto)
				}
			}
		}

		// TODO: inject errors into watches

		suite.Run("Watch", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				writer(context.Background(), doWrite(makeID(3)))

				watcher, err := watchRead.Watch()
				// Difference between postgres and etcd - etcd will error at the `Watch` call, postgres will error in the channel
				if err != nil {
					require.True(t, errors.Is(err, context.Canceled))
					return
				}
				defer watcher.Close()

				events := []*watch.Event{}
				collectEventsChannel(watcher, 1, &events)
				checkEvents(t, events, []testEvent{{watch.EventError, "", nil}})
				require.YesError(t, events[0].Err)
				require.True(t, errors.Is(events[0].Err, context.Canceled))
			})

			subsuite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(4)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				watcher, err := watchRead.Watch()
				require.NoError(t, err)
				defer watcher.Close()

				events := []*watch.Event{}
				for {
					events = append(events, <-watcher.Watch())
					if events[len(events)-1].Type == watch.EventError {
						break
					}
					if len(events) == 2 {
						cancel()
					}
				}
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
					{watch.EventError, "", nil},
				})
				require.YesError(t, events[2].Err)
				require.True(t, errors.Is(events[2].Err, context.Canceled))
			})

			subsuite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				watcher, err := watchRead.Watch()
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			subsuite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				events := []*watch.Event{}
				watcher, err := watchRead.Watch()
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 4, &events)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
					{watch.EventDelete, idA, nil}, // TODO: deleteAll order isn't well-defined?
					{watch.EventDelete, idB, nil},
				})
			})

			subsuite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				events := []*watch.Event{}
				watcher, err := watchRead.Watch()
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
				})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doWrite(id))

				events := []*watch.Event{}
				watcher, err := watchRead.Watch()
				require.NoError(t, err)
				defer watcher.Close()
				collectEventsChannel(watcher, 2, &events)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventPut, id, makeProto(id)},
				})
			})
		})

		suite.Run("WatchF", func(subsuite *testing.T) {
			subsuite.Parallel()

			// TODO: options: filter, initial sort

			subsuite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				writer(context.Background(), doWrite(makeID(1)))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(1, &events))
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				require.Equal(t, 0, len(events))
			})

			subsuite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(func(ev *watch.Event) error {
					events = append(events, ev)
					if len(events) == 2 {
						cancel() // Stop iteration via cancel once we're out of the initial value stage
					}
					return nil
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			subsuite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				})
			})

			subsuite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(4, &events))
				require.NoError(t, err)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
					{watch.EventDelete, idA, nil}, // TODO: deleteAll order isn't well-defined?
					{watch.EventDelete, idB, nil},
				})
			})

			subsuite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idB, makeProto(idB)},
				})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doWrite(id))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(2, &events))
				require.NoError(t, err)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventPut, id, makeProto(id)},
				})
			})

			subsuite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))

				events := []*watch.Event{}
				err := watchRead.WatchF(collectEventsCallback(1, &events))
				require.NoError(t, err)
				checkEvents(t, events, []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
				})
			})
		})

		suite.Run("WatchOne", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("ErrNotFound", func(t *testing.T) {
				t.Parallel()
			})
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
			})
		})

		suite.Run("WatchOneF", func(subsuite *testing.T) {
			subsuite.Parallel()

			// TODO: options: filter, initial sort

			subsuite.Run("Interruption", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(canceledContext())
				id := makeID(1)
				writer(context.Background(), doWrite(makeID(1)))

				events := []*watch.Event{}
				err := watchRead.WatchOneF(id, collectEventsCallback(1, &events))
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				require.Equal(t, 0, len(events))
			})

			subsuite.Run("InterruptionAfterInitial", func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(ctx)
				id := makeID(1)
				writer(context.Background(), doWrite(id))
				asyncWrite(t, writer, doDelete(id))

				expected := []testEvent{
					{watch.EventPut, id, makeProto(id)},
					{watch.EventDelete, id, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(id, func(ev *watch.Event) error {
					events = append(events, ev)
					if len(events) == len(expected) {
						cancel() // Stop iteration via cancel once we're out of the initial value stage
					}
					return nil
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, context.Canceled))
				checkEvents(t, events, expected)
			})

			subsuite.Run("Delete", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDelete(idA), doDelete(idB))

				expected := []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventDelete, idA, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			subsuite.Run("DeleteAll", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doDeleteAll())

				expected := []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventDelete, idA, nil},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			subsuite.Run("Create", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				expected := []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))
				asyncWrite(t, writer, doWrite(idA), doWrite(idB))

				expected := []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})

			subsuite.Run("ErrBreak", func(t *testing.T) {
				t.Parallel()
				reader, writer := newCollection(context.Background(), t)
				watchRead := reader(context.Background())
				idA := makeID(1)
				idB := makeID(2)
				writer(context.Background(), doWrite(idA))
				writer(context.Background(), doWrite(idB))

				expected := []testEvent{
					{watch.EventPut, idA, makeProto(idA)},
				}
				events := []*watch.Event{}
				err := watchRead.WatchOneF(idA, collectEventsCallback(len(expected), &events))
				require.NoError(t, err)
				checkEvents(t, events, expected)
			})
		})
	})

	parent.Run("ReadWrite", func(suite *testing.T) {
		suite.Parallel()
		initCollection := func(t *testing.T) (col.ReadOnlyCollection, WriteCallback) {
			reader, writer := newCollection(context.Background(), t)
			require.NoError(t, writer(context.Background(), populateCollection))
			return reader(context.Background()), writer
		}

		testRollback := func(t *testing.T, cb func(rw col.ReadWriteCollection) error) error {
			readOnly, writer := initCollection(t)
			err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
				return cb(rw)
			})
			require.YesError(t, err)
			checkDefaultCollection(t, readOnly, RowDiff{})
			return err
		}

		suite.Run("Get", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)

				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					itemProto := &col.TestItem{}
					if err := rw.Get(makeID(8), itemProto); err != nil {
						return err
					}
					if itemProto.Value != originalValue {
						return errors.Errorf("Incorrect value when fetching via ReadWriteCollection.Get, expected: %s, actual: %s", originalValue, itemProto.Value)
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("AfterDelete", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)

				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					if err := rw.Get(makeID(1), testProto); err != nil {
						return err
					}

					oobID := makeID(10)
					if err := rw.Get(oobID, testProto); !col.IsErrNotFound(err) {
						return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", oobID)
					}

					rw.DeleteAll()

					for _, id := range idRange(0, defaultCollectionSize) {
						if err := rw.Get(id, testProto); !col.IsErrNotFound(err) {
							return errors.Wrapf(err, "Expected ErrNotFound for key '%s', but got", id)
						}
					}
					return nil
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: idRange(0, defaultCollectionSize)})

				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})
		})

		suite.Run("Create", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return rw.Create(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("ErrExists", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Create(testProto.ID, testProto)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: overwriteID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("CreateError", func(t *testing.T) {
					t.Parallel()
					newID := makeID(10)
					overwriteID := makeID(6)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(newID)
						if err := rw.Create(testProto.ID, testProto); err != nil {
							return err
						}
						testProto = makeProto(overwriteID)
						return rw.Create(testProto.ID, testProto)
					})
					require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: overwriteID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(makeID(10))
						if err := rw.Create(testProto.ID, testProto); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Put", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := makeProto(newID)
					return rw.Put(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := makeProto(overwriteID)
					return rw.Put(testProto.ID, testProto)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()
				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := makeProto(makeID(10))
						if err := rw.Put(testProto.ID, testProto); err != nil {
							return err
						}
						testProto = makeProto(makeID(8))
						if err := rw.Put(testProto.ID, testProto); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Update", func(subsuite *testing.T) {
			subsuite.Parallel()
			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				updateID := makeID(1)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					if err := rw.Upsert(updateID, testProto, func() error {
						if err := checkItem(t, testProto, updateID, originalValue); err != nil {
							return err
						}
						testProto.Value = changedValue
						return nil
					}); err != nil {
						return err
					}
					if err := rw.Get(updateID, testProto); err != nil {
						return err
					}
					return checkItem(t, testProto, updateID, changedValue)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{updateID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				updateID := makeID(2)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return rw.Update(updateID, testProto, func() error {
						return &TestError{}
					})
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("NotFound", func(t *testing.T) {
				t.Parallel()
				notExistsID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return rw.Update(notExistsID, testProto, func() error {
						return nil
					})
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("ErrorInCallback", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(2), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return err
						}
						return rw.Update(makeID(2), testProto, func() error {
							return &TestError{}
						})
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("UpdateError", func(t *testing.T) {
					t.Parallel()
					notExistsID := makeID(10)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(3), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return err
						}
						return rw.Update(notExistsID, testProto, func() error {
							testProto.Value = changedValue
							return nil
						})
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Update(makeID(6), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Upsert", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Insert", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return rw.Upsert(newID, testProto, func() error {
						if testProto.ID != "" {
							return errors.Errorf("Upsert of new row modified the ID, new value: %s", testProto.ID)
						}
						if testProto.Value != "" {
							return errors.Errorf("Upsert of new row modified the Value, new value: %s", testProto.Value)
						}
						testProto.ID = newID
						testProto.Value = changedValue
						return nil
					})
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Created: []string{newID}})
			})

			subsuite.Run("ErrorInCallback", func(t *testing.T) {
				t.Parallel()
				newID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return rw.Upsert(newID, testProto, func() error {
						return &TestError{}
					})
				})
				require.YesError(t, err)
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("Overwrite", func(t *testing.T) {
				t.Parallel()
				overwriteID := makeID(5)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					testProto := &col.TestItem{}
					return rw.Upsert(overwriteID, testProto, func() error {
						if testProto.ID != overwriteID {
							return errors.Errorf("Upsert of existing row does not pass through the ID, expected: %s, actual: %s", overwriteID, testProto.ID)
						}
						if testProto.Value != originalValue {
							return errors.Errorf("Upsert of existing row does not pass through the value, expected: %s, actual: %s", originalValue, testProto.Value)
						}
						testProto.Value = changedValue
						return nil
					})
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Changed: []string{overwriteID}})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("UpsertError", func(t *testing.T) {
					t.Parallel()
					existsID := makeID(3)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(6), testProto, func() error {
							return nil
						}); err != nil {
							return err
						}
						return rw.Create(existsID, testProto)
					})
					require.True(t, col.IsErrExists(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrExists{Type: collectionName, Key: existsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(6), testProto, func() error {
							return nil
						}); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserErrorInCallback", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						testProto := &col.TestItem{}
						if err := rw.Upsert(makeID(5), testProto, func() error {
							testProto.Value = changedValue
							return nil
						}); err != nil {
							return err
						}

						return rw.Upsert(makeID(6), testProto, func() error {
							return &TestError{}
						})
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("Delete", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				deleteID := makeID(3)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					return rw.Delete(deleteID)
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: []string{deleteID}})
			})

			subsuite.Run("NotExists", func(t *testing.T) {
				t.Parallel()
				notExistsID := makeID(10)
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					return rw.Delete(notExistsID)
				})
				require.YesError(t, err)
				require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
				require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				checkDefaultCollection(t, readOnly, RowDiff{})
			})

			subsuite.Run("TransactionRollback", func(subsuite *testing.T) {
				subsuite.Parallel()

				subsuite.Run("DeleteError", func(t *testing.T) {
					t.Parallel()
					notExistsID := makeID(10)
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return err
						}
						return rw.Delete(notExistsID)
					})
					require.True(t, col.IsErrNotFound(err), "Incorrect error: %v", err)
					require.True(t, errors.Is(err, col.ErrNotFound{Type: collectionName, Key: notExistsID}), "Incorrect error: %v", err)
				})

				subsuite.Run("UserError", func(t *testing.T) {
					t.Parallel()
					err := testRollback(t, func(rw col.ReadWriteCollection) error {
						if err := rw.Delete(makeID(6)); err != nil {
							return err
						}
						return &TestError{}
					})
					require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
				})
			})
		})

		suite.Run("DeleteAll", func(subsuite *testing.T) {
			subsuite.Parallel()

			subsuite.Run("Success", func(t *testing.T) {
				t.Parallel()
				readOnly, writer := initCollection(t)
				err := writer(context.Background(), func(rw col.ReadWriteCollection) error {
					return rw.DeleteAll()
				})
				require.NoError(t, err)
				checkDefaultCollection(t, readOnly, RowDiff{Deleted: idRange(0, 10)})
				count, err := readOnly.Count()
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			})

			subsuite.Run("TransactionRollback", func(t *testing.T) {
				t.Parallel()
				err := testRollback(t, func(rw col.ReadWriteCollection) error {
					if err := rw.DeleteAll(); err != nil {
						return err
					}
					return &TestError{}
				})
				require.True(t, errors.Is(err, TestError{}), "Incorrect error: %v", err)
			})
		})
	})
}

// test indexes
// test multiple changes to the same row in one txn
// test interruption