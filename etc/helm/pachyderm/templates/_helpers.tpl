{{- /*
SPDX-FileCopyrightText: Pachyderm, Inc. <info@pachyderm.com>
SPDX-License-Identifier: Apache-2.0
*/ -}}
{{- /* vim: set filetype=mustache: */ -}}

{{- define "pachyderm.storageBackend" -}}
{{- if eq .Values.deployTarget "" }}
{{ fail "deployTarget must be set" }}
{{- end }}
{{- if .Values.pachd.storage.backend -}}
{{ .Values.pachd.storage.backend }}
{{- else if eq .Values.deployTarget "AMAZON" -}}
AMAZON
{{- else if eq .Values.deployTarget "GOOGLE" -}}
GOOGLE
{{- else if eq .Values.deployTarget "MICROSOFT" -}}
MICROSOFT
{{- else if eq .Values.deployTarget "LOCAL" -}}
LOCAL
{{- else -}}
{{ fail "pachd.storage.backend required when no matching deploy target found" }}
{{- end -}}
{{- end -}}

{{- define "pachyderm.consoleSecret" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.console.config.oauthClientSecret "hashSalt" "consoleSecret" "OnlyDefault" .Release.IsUpgrade) }}
{{- end -}}

{{- define "pachyderm.clusterDeploymentId" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.pachd.clusterDeploymentID "hashSalt" "deploymentID" "OnlyDefault" .Release.IsUpgrade) }}
{{- end -}}

{{- define "pachyderm.enterpriseSecret" -}}
{{ include "defaultOrStableHash" (dict "defaultValue" .Values.pachd.enterpriseSecret "hashSalt" "enterpriseSecret" "OnlyDefault" .Release.IsUpgrade) }}
{{- end -}}

## if 'defaultValue' isn't defined use the date/time to create a hash
## truncate the date to the minute and use 'genPrefix' to produce a different hash per useage
## expects a context containing 'defaultValue' and 'hashSalt' keys
{{- define "defaultOrStableHash" -}}
{{- if .defaultValue }}
    {{- .defaultValue }}
{{- else }}
    {{- if .OnlyDefault }}
        {{- fail "must provide a default value during release upgrade." }}
    {{- else }}
        {{- derivePassword 1 "long" (now | toString | trunc 16) .hashSalt "pachyderm" -}}
    {{- end }}
{{- end }}
{{- end -}}