# Mapping of NBMP descriptors

The following table gives an overview of currently supported NBMP descriptors as well as the mapping to internal nagare media engine Kubernetes resources.

🟢 Supported or support is currently implemented<br>
🟠 Partially Supported (see comment)<br>
🔴 Not supported

| Supported | NBMP Descriptor | Mapping | Comment |
| :-------: | --------------- | ------- | ------- |
| 🟢 | `scheme:` | - | |
| 🟢 | &emsp; `uri: URI` | constant | |
| 🟢 | `general: &general` | - | |
| 🟢 | &emsp; `id: string` | **Workflow**<br>`.metadata.name`<br><br>**Task**<br>`.metadata.name` | |
| 🟢 | &emsp; `name: string` | **Workflow**<br>`.spec.humanReadable.name`<br><br>**Task**<br>`.spec.humanReadable.name` | |
| 🟢 | &emsp; `description: string` | **Workflow**<br>`.spec.humanReadable.description`<br><br>**Task**<br>`.spec.humanReadable.description` | |
| 🔴 | &emsp; `rank: number` | - | only applicable to function(s) (groups) |
| 🟢 | &emsp; `nbmp-brand: URI` | constant | |
| 🟢 | &emsp; `published-time: time` | **Workflow**<br>`.metadata.creationTimestamp`<br><br>**Task**<br>`.metadata.creationTimestamp` | |
| 🟢 | &emsp; `priority: number` | **Function**<br>`.spec.template.spec.template.spec.priorityClassName`<br>`.spec.template.spec.template.spec.priority`<br><br>**Task**<br>`.spec.templatePatches.spec.template.spec.priorityClassName`<br>`.spec.templatePatches.spec.template.spec.priority`<br><br>**TaskTemplate**<br>`.spec.templatePatches.spec.template.spec.priorityClassName`<br>`.spec.templatePatches.spec.template.spec.priority` | only supported for tasks (processing.function-restrictions[*].general.priority)<br>There should be a mapping of priority -> priorityClassName in the nagare media engine configuration |
| 🔴 | &emsp; `location: string` | | ??? |
| 🔴 | &emsp; `task-group:` | | no group support for now |
| 🔴 | &emsp; &emsp; `- group-id: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `task-id:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `- string` | | |
| 🔴 | &emsp; &emsp; &emsp; `group-type: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|distance` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|sync` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|virtual` | | |
| 🔴 | &emsp; &emsp; &emsp; `group-mode: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|synchronous` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|asynchronous` | | |
| 🔴 | &emsp; &emsp; &emsp; `net-zero: bool` | | |
| 🟢 | &emsp; `input-ports:` | - | logical input ports |
| 🟢 | &emsp; &emsp; `- port-name: string` | **Task**<br>`.spec.inputs[].portBindings[].id` | |
| 🟢 | &emsp; &emsp; &emsp; `bind:` | - | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `stream-id: string` | **Task**<br>`.spec.inputs[].id` | mapping to correct input task |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `name: string` | | now optional |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `keywords:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- string` | | |
| 🟢 | &emsp; `output-ports:` | - | logical output ports |
| 🟢 | &emsp; &emsp; `- port-name: string` | **Task**<br>`.spec.outputs[].portBindings[].id` | |
| 🟢 | &emsp; &emsp; &emsp; `bind:` | - | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `stream-id: string` | **Task**<br>`.spec.outputs[].id` | mapping to correct output for task |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `name: string` | | now optional |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `keywords:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- string` | | |
| 🔴 | &emsp; `is-group: bool` | | |
| 🔴 | &emsp; `nonessential: bool` | | |
| 🟢 | &emsp; `state: \|` | - | |
| 🟢 | &emsp; &emsp; `\|instantiated` | **Workflow**<br>`.status.phase == Initializing`<br><br>**Task**<br>`.status.phase == Initializing \|\| JobPending` | |
| 🔴 | &emsp; &emsp; `\|idle` | | |
| 🟢 | &emsp; &emsp; `\|running` | **Workflow**<br>`.status.phase == Running \|\| AwaitingCompletion`<br><br>**Task**<br>`.status.phase == Running` | |
| 🟢 | &emsp; &emsp; `\|in-error` | **Workflow**<br>`.status.phase == Failed`<br><br>**Task**<br>`.status.phase == Failed` | |
| 🟠 | &emsp; &emsp; `\|destroyed` | **Workflow**<br>`.metadata.deletionTimestamp != nil`<br><br>**Task**<br>`.metadata.deletionTimestamp != nil` | |
| 🔴 | `repository:` | | |
| 🔴 | &emsp; `mode: \|` | | |
| 🔴 | &emsp; &emsp; `\|strict` | | |
| 🔴 | &emsp; &emsp; `\|preferred` | | |
| 🔴 | &emsp; &emsp; `\|available` | | |
| 🔴 | &emsp; `location:` | | |
| 🔴 | &emsp; &emsp; `- url: URI` | | |
| 🔴 | &emsp; &emsp; &emsp; `name: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `description: string` | | |
| 🟢 | `input: &input` | - | |
| 🟢 | &emsp; `media-parameters: &media-parameters` | - | |
| 🟢 | &emsp; &emsp; `stream-id: string` | **Workflow**<br>`.spec.inputs[].id`<br><br>**Task**<br>`.spec.inputs[].id` | |
| 🟢 | &emsp; &emsp; `name: string` | **Workflow**<br>`.spec.inputs[].humanReadable.name`<br><br>**Task**<br>`.spec.inputs[].humanReadable.name` | |
| 🟢 | &emsp; &emsp; `keywords:` | - | |
| 🟢 | &emsp; &emsp; &emsp; `- string1=string2` | **Workflow**<br>`.spec.inputs[].labels[string1] = string2`<br><br>**Task**<br>`.spec.inputs[].labels[string1] = string2` | |
| 🟢 | &emsp; &emsp; &emsp; `- string` | **Workflow**<br>`.spec.inputs[].labels[string] = ""`<br><br>**Task**<br>`.spec.inputs[].labels[string] = ""` | |
| 🟢 | &emsp; &emsp; `mime-type: string` | **Workflow**<br>`.spec.inputs[].metadata.mimeType`<br><br>**Task**<br>`.spec.inputs[].metadata.mimeType` | |
| 🟢 | &emsp; &emsp; `video-format:` | - | |
| 🟢 | &emsp; &emsp; &emsp; `- &parameter` | - | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `name: string` | **Workflow**<br>`.spec.inputs[].metadata.streams[].properties[].name`<br><br>**Task**<br>`.spec.inputs[].metadata.streams[].properties[].name` | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `id: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `discription: string` | | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `datatype: \|` | - | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|boolean` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|integer` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|number` | | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|string` | - | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|array` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `conditions:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `exclusions:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- number` | | |
| 🟠 | &emsp; &emsp; &emsp; &emsp; `values:` | - | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- name: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `id: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `restrictions: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|bool` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|min-value: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `max-value: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `increment: number` | | |
| 🟠 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|- string` | **Workflow**<br>`.spec.inputs[].metadata.streams[].properties[].value`<br><br>**Task**<br>`.spec.inputs[].metadata.streams[].properties[].value` | only support one value |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `schema:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `<string>: object` | | |
| 🟢 | &emsp; &emsp; `audio-format:` | - | |
| 🟢 | &emsp; &emsp; &emsp; `- *parameter` | - | |
| 🟢 | &emsp; &emsp; `image-format:` | - | |
| 🟢 | &emsp; &emsp; &emsp; `- *parameter` | - | |
| 🟢 | &emsp; &emsp; `codec-type: string` | **Workflow**<br>`.spec.inputs[].metadata.codecType`<br><br>**Task**<br>`.spec.inputs[].metadata.codecType` | |
| 🔴 | &emsp; &emsp; `protocol: string` | | |
| 🟢 | &emsp; &emsp; `mode: \|` | **Workflow**<br>`.spec.inputs[].direction`<br><br>**Task**<br>`.spec.inputs[].direction` | |
| 🟢 | &emsp; &emsp; &emsp; `\|push` | - | |
| 🟢 | &emsp; &emsp; &emsp; `\|pull` | - | |
| 🔴 | &emsp; &emsp; `throughput: number` | | |
| 🔴 | &emsp; &emsp; `buffersize: number` | | |
| 🔴 | &emsp; &emsp; `availability-duration: number` | | |
| 🔴 | &emsp; &emsp; `timeout: number` | | |
| 🟢 | &emsp; &emsp; `caching-server-url: URI` | **Workflow**<br>`.spec.inputs[].url`<br><br>**Task**<br>`.spec.inputs[].url` | |
| 🔴 | &emsp; &emsp; `completion-timeout: number` | | |
| 🟢 | &emsp; `metadata-parameters: &metadata-parameters` | - | |
| 🟢 | &emsp; &emsp; `stream-id: string` | **Workflow**<br>`.spec.inputs[].id`<br><br>**Task**<br>`.spec.inputs[].id` | |
| 🟢 | &emsp; &emsp; `name: string` | **Workflow**<br>`.spec.inputs[].humanReadable.name`<br><br>**Task**<br>`.spec.inputs[].humanReadable.name` | |
| 🟢 | &emsp; &emsp; `keywords:` | - | |
| 🟢 | &emsp; &emsp; &emsp; `- string1=string2` | **Workflow**<br>`.spec.inputs[].labels[string1] = string2`<br><br>**Task**<br>`.spec.inputs[].labels[string1] = string2` | |
| 🟢 | &emsp; &emsp; &emsp; `- string` | **Workflow**<br>`.spec.inputs[].labels[string] = ""`<br><br>**Task**<br>`.spec.inputs[].labels[string] = ""` | |
| 🟢 | &emsp; &emsp; `mime-type: string` | **Workflow**<br>`.spec.inputs[].metadata.mimeType`<br><br>**Task**<br>`.spec.inputs[].metadata.mimeType` | |
| 🟢 | &emsp; &emsp; `codec-type: string` | **Workflow**<br>`.spec.inputs[].metadata.codecType`<br><br>**Task**<br>`.spec.inputs[].metadata.codecType` | |
| 🔴 | &emsp; &emsp; `protocol: string` | | |
| 🟢 | &emsp; &emsp; `mode: \|` | **Workflow**<br>`.spec.inputs[].direction`<br><br>**Task**<br>`.spec.inputs[].direction` | |
| 🟢 | &emsp; &emsp; &emsp; `\|push` | - | |
| 🟢 | &emsp; &emsp; &emsp; `\|pull` | - | |
| 🔴 | &emsp; &emsp; `max-size: number` | | |
| 🔴 | &emsp; &emsp; `min-interval: number` | | |
| 🔴 | &emsp; &emsp; `availability-duration: number` | | |
| 🔴 | &emsp; &emsp; `timeout: number` | | |
| 🟢 | &emsp; &emsp; `caching-server-url: URI` | **Workflow**<br>`.spec.inputs[].url`<br><br>**Task**<br>`.spec.inputs[].url` | |
| 🔴 | &emsp; &emsp; `scheme-uri: URI` | | |
| 🔴 | &emsp; &emsp; `completion-timeout: number` | | |
| 🟢 | `output: &output` | - | |
| 🟢 | &emsp; `media-parameters: *media-parameters` | - | |
| 🟢 | &emsp; `metadata-parameters: *metadata-parameters` | - | |
| 🟢 | `processing: &processing` | - | |
| 🟢 | &emsp; `keywords:` | - | |
| 🟢 | &emsp; &emsp; `- string1=string2` | **Task**<br>`.spec.functionSelector.matchExpressions.key = string1`<br>`.spec.functionSelector.matchExpressions.operator = In`<br>`.spec.functionSelector.matchExpressions.key = [string2]` | |
| 🟢 | &emsp; &emsp; `- string` | **Task**<br>`.spec.functionSelector.matchExpressions.key = string`<br>`.spec.functionSelector.matchExpressions.operator = Exists` | |
| 🔴 | &emsp; `image:` | | |
| 🔴 | &emsp; &emsp; `- is-dynamic: bool` | | |
| 🔴 | &emsp; &emsp; &emsp; `url: URI` | | |
| 🔴 | &emsp; &emsp; &emsp; `static-image-info:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `os: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `version: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `architecture: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `environment: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `patch-url: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `patch-script:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `<string>: object` | | |
| 🔴 | &emsp; &emsp; &emsp; `dynamic-image-info:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `scheme: URI` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `information:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `<string>: object` | | |
| 🟢 | &emsp; `start-time: Time` | **Workflow**<br>`.status.startTime`<br><br>**Task**<br>`.status.startTime` | |
| 🟢 | &emsp; `connection-map:` | - | |
| 🟢 | &emsp; &emsp; `- connection-id: string` | dynamic value | |
| 🟢 | &emsp; &emsp; &emsp; `from: &connection-mapping-port` | - | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `id: string` | **Task**<br>`.status.functionRef.name` | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `instance: string` | **Task**<br>`.metadata.name` | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `port-name: string` | TODO | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `input-restrictions: *input` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `output-restrictions: *output` | | |
| 🟢 | &emsp; &emsp; &emsp; `to: *connection-mapping-port` | TODO | |
| 🔴 | &emsp; &emsp; &emsp; `flowcontrol: *flow-control-requirement` | | |
| 🔴 | &emsp; &emsp; &emsp; `co-located: bool` | | |
| 🔴 | &emsp; &emsp; &emsp; `breakable: bool` | | |
| 🔴 | &emsp; &emsp; &emsp; `other-parameters:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `- *parameter` | | |
| 🟢 | &emsp; `function-restrictions:` | - | |
| 🟢 | &emsp; &emsp; `- instance: string` | **Task**<br>`.metadata.name` | |
| 🟢 | &emsp; &emsp; &emsp; `general: *general` | - | |
| 🔴 | &emsp; &emsp; &emsp; `processing: *processing` | | |
| 🟠 | &emsp; &emsp; &emsp; `requirements: *requirements` | - | |
| 🟢 | &emsp; &emsp; &emsp; `configuration:` | - | |
| 🟢 | &emsp; &emsp; &emsp; &emsp; `- *parameter` | **Workflow**<br>`.spec.config`<br><br>**Task**<br>`.spec.config`<br><br><br>**TaskTemplate**<br>`.spec.config`<br><br><br>**Function**<br>`.spec.defaultConfig` | |
| 🔴 | &emsp; &emsp; &emsp; `client-assistant: *client-assistant` | | |
| 🟢 | &emsp; &emsp; &emsp; `failover: *failover` | - | |
| 🔴 | &emsp; &emsp; &emsp; `monitoring: *monitoring` | | |
| 🔴 | &emsp; &emsp; &emsp; `reporting: *reporting` | | |
| 🔴 | &emsp; &emsp; &emsp; `notification: *notification` | | |
| 🔴 | &emsp; &emsp; &emsp; `step: *step` | | |
| 🔴 | &emsp; &emsp; &emsp; `security: *security` | | |
| 🔴 | &emsp; &emsp; &emsp; `blacklist:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `- \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|requirement` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|client-assistant` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|fail-over` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|monitoring` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|reporting` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|notification` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|security` | | |
| 🟠 | `requirements: &requirements` | - | |
| 🔴 | &emsp; `flowcontrol: &flow-control-requirement` | | |
| 🔴 | &emsp; &emsp; `typical-delay: number` | | |
| 🔴 | &emsp; &emsp; `min-delay: number` | | |
| 🔴 | &emsp; &emsp; `max-delay: number` | | |
| 🔴 | &emsp; &emsp; `min-throughput: number` | | |
| 🔴 | &emsp; &emsp; `max-throughput: number` | | |
| 🔴 | &emsp; &emsp; `averaging-window: number` | | |
| 🟠 | &emsp; `hardware:` | TODO | |
| 🟠 | &emsp; &emsp; `vcpu: number` | TODO | |
| 🟠 | &emsp; &emsp; `vgpu: number` | TODO | |
| 🟠 | &emsp; &emsp; `ram: number` | TODO | |
| 🟠 | &emsp; &emsp; `disk: number` | TODO | |
| 🟠 | &emsp; &emsp; `placement: string (^[A-Z]{2}$)\|(^[A-Z]{2}-.*)` | TODO | |
| 🔴 | &emsp; `security:` | | |
| 🔴 | &emsp; &emsp; `tls: bool` | | |
| 🔴 | &emsp; &emsp; `ipsec: bool` | | |
| 🔴 | &emsp; &emsp; `cenc: bool` | | |
| 🔴 | &emsp; `workflow-task:` | | |
| 🔴 | &emsp; &emsp; `function-fusible: bool` | | |
| 🔴 | &emsp; &emsp; `function-enhancable: bool` | | |
| 🟠 | &emsp; &emsp; `execution-mode: \|` | TODO | |
| 🟠 | &emsp; &emsp; &emsp; `\|streaming` | TODO | |
| 🟠 | &emsp; &emsp; &emsp; `\|step` | TODO | |
| 🟠 | &emsp; &emsp; &emsp; `\|hybrid` | TODO | |
| 🔴 | &emsp; &emsp; `proximity:` | | |
| 🔴 | &emsp; &emsp; &emsp; `other-task-id: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `distance: number` | | |
| 🔴 | &emsp; &emsp; `proximity-equation:` | | |
| 🔴 | &emsp; &emsp; &emsp; `distance-parameters:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `- &variable` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `name: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `definition: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `unit: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `var-type: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|integer` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|float` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|boolean` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; &emsp; `\|number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `value: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `min: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `max: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `url: URI` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `children: *variable` | | |
| 🔴 | &emsp; &emsp; &emsp; `distance-equation: string` | | |
| 🔴 | &emsp; &emsp; `split-efficiency:` | | |
| 🔴 | &emsp; &emsp; &emsp; `split-norm: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|pnorm` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `\|custom` | | |
| 🔴 | &emsp; &emsp; &emsp; `split-equation: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `split-result: number` | | |
| 🔴 | &emsp; `resource-estimators:` | | |
| 🔴 | &emsp; &emsp; `default-values:` | | |
| 🔴 | &emsp; &emsp; &emsp; `- name: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `value: string` | | |
| 🔴 | &emsp; &emsp; `computational-estimator: string` | | |
| 🔴 | &emsp; &emsp; `memory-estimator: string` | | |
| 🔴 | &emsp; &emsp; `bandwidth-estimator: string` | | |
| 🔴 | `step: &step` | | This is underspecified in old spec. Probably speaks of splitting processing into multiple subtasks based on time/space. |
| 🔴 | &emsp; `step-mode: \|` | | |
| 🔴 | &emsp; &emsp; `\|stream` | | |
| 🔴 | &emsp; &emsp; `\|stateful` | | |
| 🔴 | &emsp; &emsp; `\|stateless` | | |
| 🔴 | &emsp; `variable-duration: bool` | | |
| 🔴 | &emsp; `segment-duration: number` | | |
| 🔴 | &emsp; `segment-location: bool` | | |
| 🔴 | &emsp; `segment-sequence: bool` | | |
| 🔴 | &emsp; `segment-metadata-supported-formats:` | | |
| 🔴 | &emsp; &emsp; `- \|` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|nbmp-location-bytestream-2022` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|nbmp-sequence-bytestream-2022` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|nbmp-location-json-2022` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|nbmp-sequence-json-2022` | | |
| 🔴 | &emsp; `operating-units: number` | | |
| 🔴 | &emsp; `temporal-overlap: number` | | |
| 🔴 | &emsp; `number-of-dimensions: number` | | |
| 🔴 | &emsp; `higher-dimension-segment-divisors:` | | |
| 🔴 | &emsp; &emsp; `- number` | | |
| 🔴 | &emsp; `higher-dimensions-descriptions:` | | |
| 🔴 | &emsp; &emsp; `- \|` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|width` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|height` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|RGB` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|depth` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|YUV` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|V-PCC` | | |
| 🔴 | &emsp; `higher-dimensions-segment-order:` | | |
| 🔴 | &emsp; &emsp; `- number` | | |
| 🔴 | &emsp; `higher-dimension-overlap:` | | |
| 🔴 | &emsp; &emsp; `- number` | | |
| 🔴 | &emsp; `higher-dimension-operation-units:` | | |
| 🔴 | &emsp; &emsp; `- number` | | |
| 🔴 | `client-assistant: &client-assistant` | | |
| 🔴 | &emsp; `client-assistance-flag: bool` | | |
| 🔴 | &emsp; `measurement-collection-list:` | | |
| 🔴 | &emsp; &emsp; `<string>: object` | | |
| 🔴 | &emsp; `source-assistance-information:` | | |
| 🔴 | &emsp; &emsp; `<string>: object` | | |
| 🟢 | `failover: &failover` | - | |
| 🟢 | &emsp; `failover-mode: \|` | TODO | |
| 🟠 | &emsp; &emsp; `\|restart-immediately` | TODO | Ignore for now. Kubernetes has exponential back-off. |
| 🟠 | &emsp; &emsp; `\|restart-with-delay` | TODO | Ignore for now. Kubernetes has exponential back-off. |
| 🟢 | &emsp; &emsp; `\|continue-with-last-good-state` | TODO | |
| 🔴 | &emsp; &emsp; `\|execute-backup-deployment` | | |
| 🟢 | &emsp; &emsp; `\|exit` | TODO | |
| 🟠 | &emsp; `failover-delay: number` | - | Ignore for now. Kubernetes has exponential back-off. |
| 🔴 | &emsp; `backup-deployment-url: URI` | | |
| 🔴 | &emsp; `persistence-url: URI` | | TODO |
| 🔴 | &emsp; `persistence-interval: number` | | TODO |
| 🔴 | `monitoring: &monitoring` | | |
| 🔴 | &emsp; `event:` | | |
| 🟢 | &emsp; &emsp; `- &event` | TODO | Support only in reports. |
| 🟢 | &emsp; &emsp; &emsp; `name: string` | TODO | Human readable name. |
| 🟢 | &emsp; &emsp; &emsp; `definition: string` | TODO | Human readable description. |
| 🟢 | &emsp; &emsp; &emsp; `url: URI` | TODO | Identifier as used in CloudEvents. |
| 🔴 | &emsp; `variable:` | | |
| 🔴 | &emsp; &emsp; `- *variable` | | |
| 🔴 | &emsp; `system-events:` | | |
| 🔴 | &emsp; &emsp; `- <string>: object` | | |
| 🔴 | &emsp; `system-variables:` | | |
| 🔴 | &emsp; &emsp; `- <string>: object` | | |
| 🔴 | `assertion:` | | |
| 🔴 | &emsp; `min-priority: number` | | |
| 🔴 | &emsp; `min-priority-action: \|` | | |
| 🔴 | &emsp; &emsp; `\|rebuild` | | |
| 🔴 | &emsp; &emsp; `\|restart` | | |
| 🔴 | &emsp; &emsp; `\|wait` | | |
| 🔴 | &emsp; `support-verification: bool` | | |
| 🔴 | &emsp; `verification-acknowledgement: string` | | |
| 🔴 | &emsp; `certificate: string` | | |
| 🔴 | &emsp; `assertion:` | | |
| 🔴 | &emsp; &emsp; `- name: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `value-predicate:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `evaluation-condition: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|quality` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|computational` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|input` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|output` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `check-value:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `<string>: object` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `aggregation: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|sum` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|min` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|max` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|avg` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `offset: string` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `priority: number` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `action: \|` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|rebuild` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|restart` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `\|wait` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; `action-parameters:` | | |
| 🔴 | &emsp; &emsp; &emsp; &emsp; &emsp; `- string` | | |
| 🟢 | `reporting: &reporting` | TODO | Only support in task layer. |
| 🟢 | &emsp; `event:` | TODO | |
| 🟢 | &emsp; &emsp; `- *event` | TODO | List of events to report specific to this task. |
| 🔴 | &emsp; `variable:` | | |
| 🔴 | &emsp; &emsp; `- *variable` | | |
| 🟠 | &emsp; `system-events:` | TODO | List of generic nagare events to report. |
| 🟠 | &emsp; &emsp; `- <string>: object` | TODO | Also in the *event format!<br>All system events should be reported |
| 🔴 | &emsp; `system-variables:` | | |
| 🔴 | &emsp; &emsp; `- <string>: object` | | |
| 🟢 | &emsp; `report-type: string` | constant | Value should be something with nagare. |
| 🟠 | &emsp; `reporting-interval: number` | - | Ignore as the report should be event based. |
| 🟠 | &emsp; `report-start-time: Time` | - | Ignore as the report should be event based. |
| 🟢 | &emsp; `url: URI` | - | This should point to the local task helper. |
| 🟢 | &emsp; `delivery-method: \|` | - | |
| 🟢 | &emsp; &emsp; `\|HTTP POST` | - | |
| 🔴 | `notification: &notification` | | |
| 🔴 | &emsp; `event:` | | |
| 🔴 | &emsp; &emsp; `- *event` | | |
| 🔴 | &emsp; `variable:` | | |
| 🔴 | &emsp; &emsp; `- *variable` | | |
| 🔴 | &emsp; `system-events:` | | |
| 🔴 | &emsp; &emsp; `- <string>: object` | | |
| 🔴 | &emsp; `system-variables:` | | |
| 🔴 | &emsp; &emsp; `- <string>: object` | | |
| 🔴 | &emsp; `notification-time: Time` | | |
| 🔴 | &emsp; `severity-level: string` | | |
| 🔴 | &emsp; `notification-type:` | | |
| 🔴 | &emsp; &emsp; `- \|` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|congestion` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|application` | | |
| 🔴 | &emsp; &emsp; &emsp; `\|system` | | |
| 🔴 | &emsp; `urls:` | | |
| 🔴 | &emsp; &emsp; `- URI` | | |
| 🔴 | &emsp; `notification-interval: number` | | |
| 🟢 | `acknowledge:` | - | |
| 🟢 | &emsp; `status: \|` | - | |
| 🟢 | &emsp; &emsp; `\|fulfilled` | - | |
| 🟢 | &emsp; &emsp; `\|failed` | - | |
| 🟢 | &emsp; &emsp; `\|not-supported` | - | |
| 🟢 | &emsp; &emsp; `\|partially-fulfilled` | - | |
| 🟢 | &emsp; `unsupported:` | - | |
| 🟢 | &emsp; &emsp; `- string` | - | |
| 🟢 | &emsp; `failed:` | - | |
| 🟢 | &emsp; &emsp; `- string` | - | |
| 🟢 | &emsp; `partial:` | - | |
| 🟢 | &emsp; &emsp; `- string` | - | |
| 🔴 | `security: &security` | | |
| 🔴 | &emsp; `name: string` | | |
| 🔴 | &emsp; `scope: \|` | | |
| 🔴 | &emsp; &emsp; `\|data` | | |
| 🔴 | &emsp; &emsp; `\|function` | | |
| 🔴 | &emsp; &emsp; `\|task` | | |
| 🔴 | &emsp; `authentication-method: string` | | |
| 🔴 | &emsp; `authority-url: URI` | | |
| 🔴 | &emsp; `certificate: string` | | |
| 🔴 | &emsp; `auth-token: string` | | |
| 🔴 | &emsp; `client-grants: string` | | |
| 🔴 | &emsp; `auth-token-expires: Time` | | |
| 🔴 | &emsp; `auth-token-renew: string` | | |
| 🔴 | &emsp; `auth-token-rotation: bool` | | |
| 🔴 | `scale:` | | |
| 🔴 | &emsp; `id: string` | | |
| 🔴 | &emsp; `description: string` | | |
| 🔴 | &emsp; `scaling-type: \|` | | |
| 🔴 | &emsp; &emsp; `\|MPE` | | |
| 🔴 | &emsp; &emsp; `\|split-merge` | | |
| 🔴 | &emsp; `scaling-factor: number` | | |
| 🔴 | &emsp; `status: \|` | | |
| 🔴 | &emsp; &emsp; `\|capabilities` | | |
| 🔴 | &emsp; &emsp; `\|consider` | | |
| 🔴 | &emsp; &emsp; `\|request` | | |
| 🔴 | &emsp; &emsp; `\|passed` | | |
| 🔴 | &emsp; &emsp; `\|failed` | | |
| 🔴 | &emsp; `target-id: string` | | |
| 🔴 | `schedule:` | | |
| 🔴 | &emsp; `id: string` | | |
| 🔴 | &emsp; `description: string` | | |
| 🔴 | &emsp; `schedule-type: \|` | | |
| 🔴 | &emsp; &emsp; `\|duration` | | |
| 🔴 | &emsp; &emsp; `\|segment` | | |
| 🔴 | &emsp; `schedule-table:` | | |
| 🔴 | &emsp; &emsp; `- task-id: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `start-time: string` | | |
| 🔴 | &emsp; &emsp; &emsp; `duration: number` | | |
| 🔴 | &emsp; &emsp; &emsp; `timescale: number` | | |
| 🔴 | &emsp; `number-of-segments:` | | |
| 🔴 | &emsp; &emsp; `- number` | | |
| 🔴 | &emsp; `loop: bool` | | |
| 🔴 | &emsp; `status: \|` | | |
| 🔴 | &emsp; &emsp; `\|capabilities` | | |
| 🔴 | &emsp; &emsp; `\|consider` | | |
| 🔴 | &emsp; &emsp; `\|request` | | |
| 🔴 | &emsp; &emsp; `\|passed` | | |
| 🔴 | &emsp; &emsp; `\|failed` | | |
