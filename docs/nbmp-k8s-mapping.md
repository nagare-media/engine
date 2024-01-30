# Mapping of NBMP descriptors

The following table gives an overview of currently supported NBMP descriptors as well as the mapping to internal nagare media engine Kubernetes resources.

🟢 Supported or support is currently implemented<br>
🟠 Partially Supported (see comment)<br>
🔴 Not supported

| Supported | NBMP Descriptor | Mapping | Comment |
| :-------: | --------------- | ------- | ------- |
| 🟢 | `scheme:` | - | |
| 🟢 |   `uri: URI` | constant | |
| 🟢 | `general: &general` | - | |
| 🟢 |   `id: string` | **Workflow**<br>`.metadata.name`<br><br>**Task**<br>`.metadata.name` | |
| 🟢 |   `name: string` | **Workflow**<br>`.spec.humanReadable.name`<br><br>**Task**<br>`.spec.humanReadable.name` | |
| 🟢 |   `description: string` | **Workflow**<br>`.spec.humanReadable.description`<br><br>**Task**<br>`.spec.humanReadable.description` | |
| 🔴 |   `rank: number` | - | only applicable to function(s) (groups) |
| 🟢 |   `nbmp-brand: URI` | constant | |
| 🟢 |   `published-time: time` | **Workflow**<br>`.metadata.creationTimestamp`<br><br>**Task**<br>`.metadata.creationTimestamp` | |
| 🟢 |   `priority: number` | **Function**<br>`.spec.template.spec.template.spec.priorityClassName`<br>`.spec.template.spec.template.spec.priority`<br><br>**Task**<br>`.spec.templatePatches.spec.template.spec.priorityClassName`<br>`.spec.templatePatches.spec.template.spec.priority`<br><br>**TaskTemplate**<br>`.spec.templatePatches.spec.template.spec.priorityClassName`<br>`.spec.templatePatches.spec.template.spec.priority` | only supported for tasks (processing.function-restrictions[*].general.priority)<br>There should be a mapping of priority -> priorityClassName in the nagare media engine configuration |
| 🔴 |   `location: string` | | ??? |
| 🔴 |   `task-group:` | | no group support for now |
| 🔴 |    `- group-id: string` | | |
| 🔴 |     `task-id:` | | |
| 🔴 |      `- string` | | |
| 🔴 |     `group-type: \|` | | |
| 🔴 |      `\|distance` | | |
| 🔴 |      `\|sync` | | |
| 🔴 |      `\|virtual` | | |
| 🔴 |     `group-mode: \|` | | |
| 🔴 |      `\|synchronous` | | |
| 🔴 |      `\|asynchronous` | | |
| 🔴 |     `net-zero: bool` | | |
| 🟢 |   `input-ports:` | - | logical input ports |
| 🟢 |    `- port-name: string` | **Task**<br>`.spec.inputs[].portBindings[].id` | |
| 🟢 |     `bind:` | - | |
| 🟢 |      `stream-id: string` | **Task**<br>`.spec.inputs[].id` | mapping to correct input task |
| 🔴 |      `name: string` | | now optional |
| 🔴 |      `keywords:` | | |
| 🔴 |       `- string` | | |
| 🟢 |   `output-ports:` | - | logical output ports |
| 🟢 |    `- port-name: string` | **Task**<br>`.spec.outputs[].portBindings[].id` | |
| 🟢 |     `bind:` | - | |
| 🟢 |      `stream-id: string` | **Task**<br>`.spec.outputs[].id` | mapping to correct output for task |
| 🔴 |      `name: string` | | now optional |
| 🔴 |      `keywords:` | | |
| 🔴 |       `- string` | | |
| 🔴 |   `is-group: bool` | | |
| 🔴 |   `nonessential: bool` | | |
| 🟢 |   `state: \|` | - | |
| 🟢 |    `\|instantiated` | **Workflow**<br>`.status.phase == Initializing`<br><br>**Task**<br>`.status.phase == Initializing \|\| JobPending` | |
| 🔴 |    `\|idle` | | |
| 🟢 |    `\|running` | **Workflow**<br>`.status.phase == Running \|\| AwaitingCompletion`<br><br>**Task**<br>`.status.phase == Running` | |
| 🟢 |    `\|in-error` | **Workflow**<br>`.status.phase == Failed`<br><br>**Task**<br>`.status.phase == Failed` | |
| 🟠 |    `\|destroyed` | **Workflow**<br>`.metadata.deletionTimestamp != nil`<br><br>**Task**<br>`.metadata.deletionTimestamp != nil` | |
| 🔴 | `repository:` | | |
| 🔴 |   `mode: \|` | | |
| 🔴 |    `\|strict` | | |
| 🔴 |    `\|preferred` | | |
| 🔴 |    `\|available` | | |
| 🔴 |   `location:` | | |
| 🔴 |    `- url: URI` | | |
| 🔴 |     `name: string` | | |
| 🔴 |     `description: string` | | |
| 🟢 | `input: &input` | - | |
| 🟢 |   `media-parameters: &media-parameters` | - | |
| 🟢 |    `stream-id: string` | **Workflow**<br>`.spec.inputs[].id`<br><br>**Task**<br>`.spec.inputs[].id` | |
| 🟢 |    `name: string` | **Workflow**<br>`.spec.inputs[].humanReadable.name`<br><br>**Task**<br>`.spec.inputs[].humanReadable.name` | |
| 🟢 |    `keywords:` | - | |
| 🟢 |     `- string1=string2` | **Workflow**<br>`.spec.inputs[].labels[string1] = string2`<br><br>**Task**<br>`.spec.inputs[].labels[string1] = string2` | |
| 🟢 |     `- string` | **Workflow**<br>`.spec.inputs[].labels[string] = ""`<br><br>**Task**<br>`.spec.inputs[].labels[string] = ""` | |
| 🟢 |    `mime-type: string` | **Workflow**<br>`.spec.inputs[].metadata.mimeType`<br><br>**Task**<br>`.spec.inputs[].metadata.mimeType` | |
| 🟢 |    `video-format:` | - | |
| 🟢 |     `- &parameter` | - | |
| 🟢 |      `name: string` | **Workflow**<br>`.spec.inputs[].metadata.streams[].properties[].name`<br><br>**Task**<br>`.spec.inputs[].metadata.streams[].properties[].name` | |
| 🔴 |      `id: number` | | |
| 🔴 |      `discription: string` | | |
| 🟢 |      `datatype: \|` | - | |
| 🔴 |       `\|boolean` | | |
| 🔴 |       `\|integer` | | |
| 🔴 |       `\|number` | | |
| 🟢 |       `\|string` | - | |
| 🔴 |       `\|array` | | |
| 🔴 |      `conditions:` | | |
| 🔴 |       `- number` | | |
| 🔴 |      `exclusions:` | | |
| 🔴 |       `- number` | | |
| 🟠 |      `values:` | - | |
| 🔴 |       `- name: string` | | |
| 🔴 |        `id: number` | | |
| 🔴 |        `restrictions: \|` | | |
| 🔴 |         `\|bool` | | |
| 🔴 |         `\|min-value: number` | | |
| 🔴 |          `max-value: number` | | |
| 🔴 |          `increment: number` | | |
| 🟠 |         `\|- string` | **Workflow**<br>`.spec.inputs[].metadata.streams[].properties[].value`<br><br>**Task**<br>`.spec.inputs[].metadata.streams[].properties[].value` | only support one value |
| 🔴 |      `schema:` | | |
| 🔴 |       `<string>: object` | | |
| 🟢 |    `audio-format:` | - | |
| 🟢 |     `- *parameter` | - | |
| 🟢 |    `image-format:` | - | |
| 🟢 |     `- *parameter` | - | |
| 🟢 |    `codec-type: string` | **Workflow**<br>`.spec.inputs[].metadata.codecType`<br><br>**Task**<br>`.spec.inputs[].metadata.codecType` | |
| 🔴 |    `protocol: string` | | |
| 🟢 |    `mode: \|` | **Workflow**<br>`.spec.inputs[].direction`<br><br>**Task**<br>`.spec.inputs[].direction` | |
| 🟢 |     `\|push` | - | |
| 🟢 |     `\|pull` | - | |
| 🔴 |    `throughput: number` | | |
| 🔴 |    `buffersize: number` | | |
| 🔴 |    `availability-duration: number` | | |
| 🔴 |    `timeout: number` | | |
| 🟢 |    `caching-server-url: URI` | **Workflow**<br>`.spec.inputs[].url`<br><br>**Task**<br>`.spec.inputs[].url` | |
| 🔴 |    `completion-timeout: number` | | |
| 🟢 |   `metadata-parameters: &metadata-parameters` | - | |
| 🟢 |    `stream-id: string` | **Workflow**<br>`.spec.inputs[].id`<br><br>**Task**<br>`.spec.inputs[].id` | |
| 🟢 |    `name: string` | **Workflow**<br>`.spec.inputs[].humanReadable.name`<br><br>**Task**<br>`.spec.inputs[].humanReadable.name` | |
| 🟢 |    `keywords:` | - | |
| 🟢 |     `- string1=string2` | **Workflow**<br>`.spec.inputs[].labels[string1] = string2`<br><br>**Task**<br>`.spec.inputs[].labels[string1] = string2` | |
| 🟢 |     `- string` | **Workflow**<br>`.spec.inputs[].labels[string] = ""`<br><br>**Task**<br>`.spec.inputs[].labels[string] = ""` | |
| 🟢 |    `mime-type: string` | **Workflow**<br>`.spec.inputs[].metadata.mimeType`<br><br>**Task**<br>`.spec.inputs[].metadata.mimeType` | |
| 🟢 |    `codec-type: string` | **Workflow**<br>`.spec.inputs[].metadata.codecType`<br><br>**Task**<br>`.spec.inputs[].metadata.codecType` | |
| 🔴 |    `protocol: string` | | |
| 🟢 |    `mode: \|` | **Workflow**<br>`.spec.inputs[].direction`<br><br>**Task**<br>`.spec.inputs[].direction` | |
| 🟢 |     `\|push` | - | |
| 🟢 |     `\|pull` | - | |
| 🔴 |    `max-size: number` | | |
| 🔴 |    `min-interval: number` | | |
| 🔴 |    `availability-duration: number` | | |
| 🔴 |    `timeout: number` | | |
| 🟢 |    `caching-server-url: URI` | **Workflow**<br>`.spec.inputs[].url`<br><br>**Task**<br>`.spec.inputs[].url` | |
| 🔴 |    `scheme-uri: URI` | | |
| 🔴 |    `completion-timeout: number` | | |
| 🟢 | `output: &output` | - | |
| 🟢 |   `media-parameters: *media-parameters` | - | |
| 🟢 |   `metadata-parameters: *metadata-parameters` | - | |
| 🟢 | `processing: &processing` | - | |
| 🟢 |   `keywords:` | - | |
| 🟢 |    `- string1=string2` | **Task**<br>`.spec.functionSelector.matchExpressions.key = string1`<br>`.spec.functionSelector.matchExpressions.operator = In`<br>`.spec.functionSelector.matchExpressions.key = [string2]` | |
| 🟢 |    `- string` | **Task**<br>`.spec.functionSelector.matchExpressions.key = string`<br>`.spec.functionSelector.matchExpressions.operator = Exists` | |
| 🔴 |   `image:` | | |
| 🔴 |    `- is-dynamic: bool` | | |
| 🔴 |     `url: URI` | | |
| 🔴 |     `static-image-info:` | | |
| 🔴 |      `os: string` | | |
| 🔴 |      `version: string` | | |
| 🔴 |      `architecture: string` | | |
| 🔴 |      `environment: string` | | |
| 🔴 |      `patch-url: string` | | |
| 🔴 |      `patch-script:` | | |
| 🔴 |       `<string>: object` | | |
| 🔴 |     `dynamic-image-info:` | | |
| 🔴 |      `scheme: URI` | | |
| 🔴 |      `information:` | | |
| 🔴 |       `<string>: object` | | |
| 🟢 |   `start-time: Time` | **Workflow**<br>`.status.startTime`<br><br>**Task**<br>`.status.startTime` | |
| 🟢 |   `connection-map:` | - | |
| 🟢 |    `- connection-id: string` | dynamic value | |
| 🟢 |     `from: &connection-mapping-port` | - | |
| 🟢 |      `id: string` | **Task**<br>`.status.functionRef.name` | |
| 🟢 |      `instance: string` | **Task**<br>`.metadata.name` | |
| 🟢 |      `port-name: string` | TODO | |
| 🔴 |      `input-restrictions: *input` | | |
| 🔴 |      `output-restrictions: *output` | | |
| 🟢 |     `to: *connection-mapping-port` | TODO | |
| 🔴 |     `flowcontrol: *flow-control-requirement` | | |
| 🔴 |     `co-located: bool` | | |
| 🔴 |     `breakable: bool` | | |
| 🔴 |     `other-parameters:` | | |
| 🔴 |      `- *parameter` | | |
| 🟢 |   `function-restrictions:` | - | |
| 🟢 |    `- instance: string` | **Task**<br>`.metadata.name` | |
| 🟢 |     `general: *general` | - | |
| 🔴 |     `processing: *processing` | | |
| 🟠 |     `requirements: *requirements` | - | |
| 🟢 |     `configuration:` | - | |
| 🟢 |      `- *parameter` | **Workflow**<br>`.spec.config`<br><br>**Task**<br>`.spec.config`<br><br><br>**TaskTemplate**<br>`.spec.config`<br><br><br>**Function**<br>`.spec.defaultConfig` | |
| 🔴 |     `client-assistant: *client-assistant` | | |
| 🟢 |     `failover: *failover` | - | |
| 🔴 |     `monitoring: *monitoring` | | |
| 🔴 |     `reporting: *reporting` | | |
| 🔴 |     `notification: *notification` | | |
| 🔴 |     `step: *step` | | |
| 🔴 |     `security: *security` | | |
| 🔴 |     `blacklist:` | | |
| 🔴 |      `- \|` | | |
| 🔴 |       `\|requirement` | | |
| 🔴 |       `\|client-assistant` | | |
| 🔴 |       `\|fail-over` | | |
| 🔴 |       `\|monitoring` | | |
| 🔴 |       `\|reporting` | | |
| 🔴 |       `\|notification` | | |
| 🔴 |       `\|security` | | |
| 🟠 | `requirements: &requirements` | - | |
| 🔴 |   `flowcontrol: &flow-control-requirement` | | |
| 🔴 |    `typical-delay: number` | | |
| 🔴 |    `min-delay: number` | | |
| 🔴 |    `max-delay: number` | | |
| 🔴 |    `min-throughput: number` | | |
| 🔴 |    `max-throughput: number` | | |
| 🔴 |    `averaging-window: number` | | |
| 🟠 |   `hardware:` | TODO | |
| 🟠 |    `vcpu: number` | TODO | |
| 🟠 |    `vgpu: number` | TODO | |
| 🟠 |    `ram: number` | TODO | |
| 🟠 |    `disk: number` | TODO | |
| 🟠 |    `placement: string (^[A-Z]{2}$)\|(^[A-Z]{2}-.*)` | TODO | |
| 🔴 |   `security:` | | |
| 🔴 |    `tls: bool` | | |
| 🔴 |    `ipsec: bool` | | |
| 🔴 |    `cenc: bool` | | |
| 🔴 |   `workflow-task:` | | |
| 🔴 |    `function-fusible: bool` | | |
| 🔴 |    `function-enhancable: bool` | | |
| 🟠 |    `execution-mode: \|` | TODO | |
| 🟠 |     `\|streaming` | TODO | |
| 🟠 |     `\|step` | TODO | |
| 🟠 |     `\|hybrid` | TODO | |
| 🔴 |    `proximity:` | | |
| 🔴 |     `other-task-id: string` | | |
| 🔴 |     `distance: number` | | |
| 🔴 |    `proximity-equation:` | | |
| 🔴 |     `distance-parameters:` | | |
| 🔴 |      `- &variable` | | |
| 🔴 |       `name: string` | | |
| 🔴 |       `definition: string` | | |
| 🔴 |       `unit: string` | | |
| 🔴 |       `var-type: \|` | | |
| 🔴 |        `\|string` | | |
| 🔴 |        `\|integer` | | |
| 🔴 |        `\|float` | | |
| 🔴 |        `\|boolean` | | |
| 🔴 |        `\|number` | | |
| 🔴 |       `value: string` | | |
| 🔴 |       `min: number` | | |
| 🔴 |       `max: number` | | |
| 🔴 |       `url: URI` | | |
| 🔴 |       `children: *variable` | | |
| 🔴 |     `distance-equation: string` | | |
| 🔴 |    `split-efficiency:` | | |
| 🔴 |     `split-norm: \|` | | |
| 🔴 |      `\|pnorm` | | |
| 🔴 |      `\|custom` | | |
| 🔴 |     `split-equation: string` | | |
| 🔴 |     `split-result: number` | | |
| 🔴 |   `resource-estimators:` | | |
| 🔴 |    `default-values:` | | |
| 🔴 |     `- name: string` | | |
| 🔴 |      `value: string` | | |
| 🔴 |    `computational-estimator: string` | | |
| 🔴 |    `memory-estimator: string` | | |
| 🔴 |    `bandwidth-estimator: string` | | |
| 🔴 | `step: &step` | | This is underspecified in old spec. Probably speaks of splitting processing into multiple subtasks based on time/space. |
| 🔴 |   `step-mode: \|` | | |
| 🔴 |    `\|stream` | | |
| 🔴 |    `\|stateful` | | |
| 🔴 |    `\|stateless` | | |
| 🔴 |   `variable-duration: bool` | | |
| 🔴 |   `segment-duration: number` | | |
| 🔴 |   `segment-location: bool` | | |
| 🔴 |   `segment-sequence: bool` | | |
| 🔴 |   `segment-metadata-supported-formats:` | | |
| 🔴 |    `- \|` | | |
| 🔴 |     `\|nbmp-location-bytestream-2022` | | |
| 🔴 |     `\|nbmp-sequence-bytestream-2022` | | |
| 🔴 |     `\|nbmp-location-json-2022` | | |
| 🔴 |     `\|nbmp-sequence-json-2022` | | |
| 🔴 |   `operating-units: number` | | |
| 🔴 |   `temporal-overlap: number` | | |
| 🔴 |   `number-of-dimensions: number` | | |
| 🔴 |   `higher-dimension-segment-divisors:` | | |
| 🔴 |    `- number` | | |
| 🔴 |   `higher-dimensions-descriptions:` | | |
| 🔴 |    `- \|` | | |
| 🔴 |     `\|width` | | |
| 🔴 |     `\|height` | | |
| 🔴 |     `\|RGB` | | |
| 🔴 |     `\|depth` | | |
| 🔴 |     `\|YUV` | | |
| 🔴 |     `\|V-PCC` | | |
| 🔴 |   `higher-dimensions-segment-order:` | | |
| 🔴 |    `- number` | | |
| 🔴 |   `higher-dimension-overlap:` | | |
| 🔴 |    `- number` | | |
| 🔴 |   `higher-dimension-operation-units:` | | |
| 🔴 |    `- number` | | |
| 🔴 | `client-assistant: &client-assistant` | | |
| 🔴 |   `client-assistance-flag: bool` | | |
| 🔴 |   `measurement-collection-list:` | | |
| 🔴 |    `<string>: object` | | |
| 🔴 |   `source-assistance-information:` | | |
| 🔴 |    `<string>: object` | | |
| 🟢 | `failover: &failover` | - | |
| 🟢 |   `failover-mode: \|` | TODO | |
| 🟠 |    `\|restart-immediately` | TODO | Ignore for now. Kubernetes has exponential back-off. |
| 🟠 |    `\|restart-with-delay` | TODO | Ignore for now. Kubernetes has exponential back-off. |
| 🟢 |    `\|continue-with-last-good-state` | TODO | |
| 🔴 |    `\|execute-backup-deployment` | | |
| 🟢 |    `\|exit` | TODO | |
| 🟠 |   `failover-delay: number` | - | Ignore for now. Kubernetes has exponential back-off. |
| 🔴 |   `backup-deployment-url: URI` | | |
| 🔴 |   `persistence-url: URI` | | TODO |
| 🔴 |   `persistence-interval: number` | | TODO |
| 🔴 | `monitoring: &monitoring` | | |
| 🔴 |   `event:` | | |
| 🟢 |    `- &event` | TODO | Support only in reports. |
| 🟢 |     `name: string` | TODO | Human readable name. |
| 🟢 |     `definition: string` | TODO | Human readable description. |
| 🟢 |     `url: URI` | TODO | Identifier as used in CloudEvents. |
| 🔴 |   `variable:` | | |
| 🔴 |    `- *variable` | | |
| 🔴 |   `system-events:` | | |
| 🔴 |    `- <string>: object` | | |
| 🔴 |   `system-variables:` | | |
| 🔴 |    `- <string>: object` | | |
| 🔴 | `assertion:` | | |
| 🔴 |   `min-priority: number` | | |
| 🔴 |   `min-priority-action: \|` | | |
| 🔴 |    `\|rebuild` | | |
| 🔴 |    `\|restart` | | |
| 🔴 |    `\|wait` | | |
| 🔴 |   `support-verification: bool` | | |
| 🔴 |   `verification-acknowledgement: string` | | |
| 🔴 |   `certificate: string` | | |
| 🔴 |   `assertion:` | | |
| 🔴 |    `- name: string` | | |
| 🔴 |     `value-predicate:` | | |
| 🔴 |      `evaluation-condition: \|` | | |
| 🔴 |       `\|quality` | | |
| 🔴 |       `\|computational` | | |
| 🔴 |       `\|input` | | |
| 🔴 |       `\|output` | | |
| 🔴 |      `check-value:` | | |
| 🔴 |       `<string>: object` | | |
| 🔴 |      `aggregation: \|` | | |
| 🔴 |       `\|sum` | | |
| 🔴 |       `\|min` | | |
| 🔴 |       `\|max` | | |
| 🔴 |       `\|avg` | | |
| 🔴 |      `offset: string` | | |
| 🔴 |      `priority: number` | | |
| 🔴 |      `action: \|` | | |
| 🔴 |       `\|rebuild` | | |
| 🔴 |       `\|restart` | | |
| 🔴 |       `\|wait` | | |
| 🔴 |      `action-parameters:` | | |
| 🔴 |       `- string` | | |
| 🟢 | `reporting: &reporting` | TODO | Only support in task layer. |
| 🟢 |   `event:` | TODO | |
| 🟢 |    `- *event` | TODO | List of events to report specific to this task. |
| 🔴 |   `variable:` | | |
| 🔴 |    `- *variable` | | |
| 🟠 |   `system-events:` | TODO | List of generic nagare events to report. |
| 🟠 |    `- <string>: object` | TODO | Also in the *event format!<br>All system events should be reported |
| 🔴 |   `system-variables:` | | |
| 🔴 |    `- <string>: object` | | |
| 🟢 |   `report-type: string` | constant | Value should be something with nagare. |
| 🟠 |   `reporting-interval: number` | - | Ignore as the report should be event based. |
| 🟠 |   `report-start-time: Time` | - | Ignore as the report should be event based. |
| 🟢 |   `url: URI` | - | This should point to the local task helper. |
| 🟢 |   `delivery-method: \|` | - | |
| 🟢 |    `\|HTTP POST` | - | |
| 🔴 | `notification: &notification` | | |
| 🔴 |   `event:` | | |
| 🔴 |    `- *event` | | |
| 🔴 |   `variable:` | | |
| 🔴 |    `- *variable` | | |
| 🔴 |   `system-events:` | | |
| 🔴 |    `- <string>: object` | | |
| 🔴 |   `system-variables:` | | |
| 🔴 |    `- <string>: object` | | |
| 🔴 |   `notification-time: Time` | | |
| 🔴 |   `severity-level: string` | | |
| 🔴 |   `notification-type:` | | |
| 🔴 |    `- \|` | | |
| 🔴 |     `\|congestion` | | |
| 🔴 |     `\|application` | | |
| 🔴 |     `\|system` | | |
| 🔴 |   `urls:` | | |
| 🔴 |    `- URI` | | |
| 🔴 |   `notification-interval: number` | | |
| 🟢 | `acknowledge:` | - | |
| 🟢 |   `status: \|` | - | |
| 🟢 |    `\|fulfilled` | - | |
| 🟢 |    `\|failed` | - | |
| 🟢 |    `\|not-supported` | - | |
| 🟢 |    `\|partially-fulfilled` | - | |
| 🟢 |   `unsupported:` | - | |
| 🟢 |    `- string` | - | |
| 🟢 |   `failed:` | - | |
| 🟢 |    `- string` | - | |
| 🟢 |   `partial:` | - | |
| 🟢 |    `- string` | - | |
| 🔴 | `security: &security` | | |
| 🔴 |   `name: string` | | |
| 🔴 |   `scope: \|` | | |
| 🔴 |    `\|data` | | |
| 🔴 |    `\|function` | | |
| 🔴 |    `\|task` | | |
| 🔴 |   `authentication-method: string` | | |
| 🔴 |   `authority-url: URI` | | |
| 🔴 |   `certificate: string` | | |
| 🔴 |   `auth-token: string` | | |
| 🔴 |   `client-grants: string` | | |
| 🔴 |   `auth-token-expires: Time` | | |
| 🔴 |   `auth-token-renew: string` | | |
| 🔴 |   `auth-token-rotation: bool` | | |
| 🔴 | `scale:` | | |
| 🔴 |   `id: string` | | |
| 🔴 |   `description: string` | | |
| 🔴 |   `scaling-type: \|` | | |
| 🔴 |    `\|MPE` | | |
| 🔴 |    `\|split-merge` | | |
| 🔴 |   `scaling-factor: number` | | |
| 🔴 |   `status: \|` | | |
| 🔴 |    `\|capabilities` | | |
| 🔴 |    `\|consider` | | |
| 🔴 |    `\|request` | | |
| 🔴 |    `\|passed` | | |
| 🔴 |    `\|failed` | | |
| 🔴 |   `target-id: string` | | |
| 🔴 | `schedule:` | | |
| 🔴 |   `id: string` | | |
| 🔴 |   `description: string` | | |
| 🔴 |   `schedule-type: \|` | | |
| 🔴 |    `\|duration` | | |
| 🔴 |    `\|segment` | | |
| 🔴 |   `schedule-table:` | | |
| 🔴 |    `- task-id: string` | | |
| 🔴 |     `start-time: string` | | |
| 🔴 |     `duration: number` | | |
| 🔴 |     `timescale: number` | | |
| 🔴 |   `number-of-segments:` | | |
| 🔴 |    `- number` | | |
| 🔴 |   `loop: bool` | | |
| 🔴 |   `status: \|` | | |
| 🔴 |    `\|capabilities` | | |
| 🔴 |    `\|consider` | | |
| 🔴 |    `\|request` | | |
| 🔴 |    `\|passed` | | |
| 🔴 |    `\|failed` | | |
