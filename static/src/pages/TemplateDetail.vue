<template>
  <v-layout column>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
          &nbsp;&nbsp;应用模板&nbsp;&nbsp;/&nbsp;&nbsp;{{ Id ? Title : '新增应用模板' }}
          <v-spacer></v-spacer>
          <v-btn icon class="green--text text--lighten-2" @click.native="exportTemplate()" title="导出">
            <v-icon light>redo</v-icon>
          </v-btn>
          <v-btn icon class="green--text text--lighten-2" @click.native="importTemplate()" title="导入">
            <v-icon light>undo</v-icon>
          </v-btn>
        </v-card-title>
        <div>
          <v-layout row justify-center>
            <v-dialog v-model="TemplateDataDlg" persistent width="640">
              <v-card>
                <v-card-row>
                  <v-card-text>
                    <v-text-field 
                      v-model="TemplateData"
                      :readonly="!Importing"
                      multi-line
                      rows="24"
                      full-width
                      class="env-list"
                    ></v-text-field>
                  </v-card-text>
                </v-card-row>
                <v-card-row actions v-if="Importing">
                  <v-btn class="blue--text darken-1" flat @click.native="doImportTemplate()">确认</v-btn>
                  <v-btn class="grey--text darken-1" flat @click.native="TemplateDataDlg = false">取消</v-btn>
                </v-card-row>
                <v-card-row actions v-if="!Importing">
                  <v-btn class="grey--text darken-1" flat @click.native="TemplateDataDlg = false">关闭</v-btn>
                </v-card-row>
              </v-card>
            </v-dialog>
          </v-layout>
          <v-layout row justify-center>
            <v-dialog v-model="EnvListDlg" persistent width="640">
              <v-card>
                <v-card-row>
                  <v-card-text>
                    <v-text-field 
                      v-model="EnvList"
                      :readonly="!Importing"
                      multi-line
                      rows="24"
                      full-width
                      class="env-list"
                    ></v-text-field>
                  </v-card-text>
                </v-card-row>
                <v-card-row actions v-if="Importing">
                  <v-btn class="blue--text darken-1" flat @click.native="doImportEnvs()">确认</v-btn>
                  <v-btn class="grey--text darken-1" flat @click.native="EnvListDlg = false">取消</v-btn>
                </v-card-row>
                <v-card-row actions v-if="!Importing">
                  <v-btn class="grey--text darken-1" flat @click.native="EnvListDlg = false">关闭</v-btn>
                </v-card-row>
              </v-card>
            </v-dialog>
          </v-layout>
          <v-layout row justify-center>
            <v-snackbar
                v-model="ApplicationNameWarning"
                timeout="1500"
                top
                error
              >应用ID改动将会影响到后续升级，请慎重</v-snackbar>
            <v-snackbar
                v-model="ServiceNameWarning"
                timeout="1500"
                top
                error
              >服务ID改动将会影响到后续升级，请慎重</v-snackbar>
            <v-snackbar
                v-model="NetworkModeWarning"
                timeout="1500"
                top
                error
              >切换网络会导致容器IP变动，可能会影响应用运行</v-snackbar>
          </v-layout>
          <v-container fluid>
            <v-layout row wrap>
              <v-flex xs2>
                <v-subheader>应用名称<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  ref="Title"
                  v-model="Title"
                  required
                  :rules="rules.Title"
                  @input="rules.Title = rules0.Title"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>应用ID<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  ref="Name"
                  v-model="Name"
                  required
                  :rules="rules.Name"
                  @input="rules.Name = rules0.Name"
                  @focus="OriginalValue = Name"
                  @change="showApplicationNameWarning"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
                <v-subheader>应用版本<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  ref="Version"
                  v-model="Version"
                  required
                  :rules="rules.Version"
                  @input="rules.Version = rules0.Version"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>说明</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Description"
                ></v-text-field>
              </v-flex>
            </v-layout>
          </v-container>
        </div>
      </v-card>
    </v-flex>
    <v-flex xs12>
      <v-card-title style="padding-left:0;">
        &nbsp;&nbsp;服务列表
        <v-spacer></v-spacer>
        <v-btn outline small class="green green--text" @click.native="addService()">
          <v-icon class="green--text">add</v-icon>添加服务
        </v-btn>
      </v-card-title>
      <div>
        <v-card v-for="(item, index) in Services" :key="item.Id" class="mb-2">
          <v-card-title>
            服务{{ index + 1 }}: {{ item.Title }}&nbsp;&nbsp;&nbsp;&nbsp;
            <span style="color:#9F9F9F;">
              域名: {{ Name }}-{{ item.Name }}.${DOMAIN_SUFFIX}
            </span>
            <v-spacer></v-spacer>
            <v-btn v-if="item.hidden" outline small icon class="blue blue--text mr-2" @click.native="hideService(item, false)" title="展开">
              <v-icon>arrow_drop_down</v-icon>
            </v-btn>
            <v-btn v-if="!item.hidden" outline small icon class="blue blue--text mr-2" @click.native="hideService(item, true)" title="折叠">
              <v-icon>arrow_drop_up</v-icon>
            </v-btn>
            <v-btn v-if="index < Services.length - 1" outline small icon class="blue blue--text mr-2" @click.native="downward(Services, index)" title="下移">
              <v-icon>arrow_downward</v-icon>
            </v-btn>
            <v-btn v-if="index > 0" outline small icon class="green green--text mr-2" @click.native="upward(Services, index)" title="上移">
              <v-icon>arrow_upward</v-icon>
            </v-btn>
            <v-btn outline small icon class="red--text text--lighten-2" @click.native="removei(Services, index)" title="删除">
              <v-icon>close</v-icon>
            </v-btn>
          </v-card-title>
          <div v-show="!item.hidden">
            <v-container fluid>
              <v-layout row wrap>
                <v-flex xs2>
                  <v-subheader>服务名称<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    :ref="'Service_Title_' + item.Id"
                    v-model="item.Title"
                    required
                    :rules="rules.Services[item.Id].Title"
                    @input="rules.Services[item.Id].Title = rules0.Services.Title"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>服务ID<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    :ref="'Service_Name_' + item.Id"
                    v-model="item.Name"
                    required
                    :rules="rules.Services[item.Id].Name"
                    @input="rules.Services[item.Id].Name = rules0.Services.Name"
                    @focus="OriginalValue = item.Name"
                    @change="showServiceNameWarning"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>镜像名称<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs5>
                  <v-text-field
                    :ref="'Service_ImageName_' + item.Id"
                    v-model="item.ImageName"
                    required
                    :rules="rules.Services[item.Id].ImageName"
                    @input="rules.Services[item.Id].ImageName = rules0.Services.ImageName"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>镜像Tag<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    :ref="'Service_ImageTag_' + item.Id"
                    v-model="item.ImageTag"
                    required
                    :rules="rules.Services[item.Id].ImageTag"
                    @input="rules.Services[item.Id].ImageTag = rules0.Services.ImageTag"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>CPU个数</v-subheader>
                </v-flex>
                <v-flex xs2>
                  <v-text-field
                    :ref="'Service_CPU_' + item.Id"
                    v-model="item.CPU"
                    placeholder="自动分配"
                    :rules="rules.Services[item.Id].CPU"
                    @input="rules.Services[item.Id].CPU = rules0.Services.CPU; if (!(item.CPU > 0)) item.ExclusiveCPU = false;"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-checkbox label="独占CPU" v-model="item.ExclusiveCPU" dark :disabled="!(item.CPU > 0)"></v-checkbox>
                </v-flex>
                <v-flex xs1>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>内存 (MB)<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    :ref="'Service_Memory_' + item.Id"
                    v-model="item.Memory"
                    required
                    :rules="rules.Services[item.Id].Memory"
                    @input="rules.Services[item.Id].Memory = rules0.Services.Memory"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>容器个数<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs2>
                  <v-text-field
                    :ref="'Service_ReplicaCount_' + item.Id"
                    v-model="item.ReplicaCount"
                    required
                    :rules="rules.Services[item.Id].ReplicaCount"
                    @input="rules.Services[item.Id].ReplicaCount = rules0.Services.ReplicaCount"
                  ></v-text-field>
                </v-flex>
                <v-flex xs3>
                  <v-checkbox label="使用宿主机网络" v-model="item.NetworkMode" true-value="host" false-value="bridge" dark @change="NetworkModeWarning = true"></v-checkbox>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>启动等待 (秒)</v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    :ref="'Service_ServiceTimeout_' + item.Id"
                    v-model="item.ServiceTimeout"
                    required
                    :rules="rules.Services[item.Id].ServiceTimeout"
                    @input="rules.Services[item.Id].ServiceTimeout = rules0.Services.ServiceTimeout"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>说明</v-subheader>
                </v-flex>
                <v-flex xs10>
                  <v-text-field
                    v-model="item.Description"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>启动命令</v-subheader>
                </v-flex>
                <v-flex xs10>
                  <v-text-field
                    v-model="item.Command"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs3>
                  <v-checkbox label="异常终止后自动重启" v-model="item.Restart" true-value="always" false-value="no" dark></v-checkbox>
                </v-flex>
                <v-flex xs7>
                </v-flex>
                <v-flex xs12 mt-5>
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>环境变量</v-subheader>
                    <v-spacer></v-spacer>
                    <v-btn icon class="green--text text--lighten-2" @click.native="exportEnvs(item)" title="导出">
                      <v-icon light>redo</v-icon>
                    </v-btn>
                    <v-btn icon class="green--text text--lighten-2" @click.native="importEnvs(item)" title="导入">
                      <v-icon light>undo</v-icon>
                    </v-btn>
                    <v-btn icon class="blue--text text--lighten-2" @click.native="addEnv(item)">
                      <v-icon light>add</v-icon>
                    </v-btn>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_envs"
                    :items="item.Envs"
                    :customSort="nosort"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.Name"
                          :ref="'Env_Name_' + props.item.index"
                          required
                          :rules="rules.Services[item.Id].Envs[props.item.Id].Name"
                          @input="rules.Services[item.Id].Envs[props.item.Id].Name = rules0.Services.Envs.Name"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Value"
                          :ref="'Env_Value_' + props.item.index"
                          required
                          :class="{ 'completer-field' : true, 'last-field': item.Envs.length == (props.item.index + 1) }"
                          :rel="'Env_Value_' + props.item.index"
                          add_func="addEnv"
                          :add_params="item.Id"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-btn icon class="red--text text--lighten-2" @click.native="removei(item.Envs, props.item.index)" title="删除">
                          <v-icon>close</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index < item.Envs.length - 1" icon class="blue--text blue--lighten-2 ml-2" @click.native="downward(item.Envs, props.item.index)" title="下移">
                          <v-icon>arrow_downward</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index > 0" icon class="green--text green--lighten-2 ml-2" @click.native="upward(item.Envs, props.item.index)" title="上移">
                          <v-icon>arrow_upward</v-icon>
                        </v-btn>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4>
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>端口申明</v-subheader>
                    <v-spacer></v-spacer>
                    <v-btn icon class="blue--text text--lighten-2" @click.native="addPort(item)">
                      <v-icon light>add</v-icon>
                    </v-btn>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_ports"
                    :items="item.Ports"
                    :customSort="nosort"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.SourcePort"
                          :ref="'Port_SourcePort_' + props.item.index"
                          required
                          :rules="rules.Services[item.Id].Ports[props.item.Id].SourcePort"
                          @input="rules.Services[item.Id].Ports[props.item.Id].SourcePort = rules0.Services.Ports.SourcePort"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.LoadBalancerId"
                          placeholder="若无需负载均衡则留空"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.TargetGroupArn"
                          placeholder="若无需负载均衡则留空"
                          :class="{ 'last-field': item.Ports.length == (props.item.index + 1) }"
                          add_func="addPort"
                          :add_params="item.Id"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-btn icon class="red--text text--lighten-2" @click.native="removei(item.Ports, props.item.index)" title="删除">
                          <v-icon>close</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index < item.Ports.length - 1" icon class="blue--text blue--lighten-2 ml-2" @click.native="downward(item.Ports, props.item.index)" title="下移">
                          <v-icon>arrow_downward</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index > 0" icon class="green--text green--lighten-2 ml-2" @click.native="upward(item.Ports, props.item.index)" title="上移">
                          <v-icon>arrow_upward</v-icon>
                        </v-btn>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4>
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>数据卷</v-subheader>
                    <v-spacer></v-spacer>
                    <v-btn icon class="blue--text text--lighten-2" @click.native="addVolumn(item)">
                      <v-icon light>add</v-icon>
                    </v-btn>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_volumns"
                    :items="item.Volumns"
                    :customSort="nosort"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.ContainerPath"
                          :ref="'Volumn_ContainerPath_' + props.item.index"
                          required
                          :rules="rules.Services[item.Id].Volumns[props.item.Id].ContainerPath"
                          @input="rules.Services[item.Id].Volumns[props.item.Id].ContainerPath = rules0.Services.Volumns.ContainerPath"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-select
                          :items="MountTypeList"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.MountType"
                          dark></v-select>
                      </td>
                      <td>
                        <v-select
                          :items="MediaTypeList"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.MediaType"
                          dark
                          @input="mediaTypeChanged(props.item)"></v-select>
                      </td>
                      <td v-if="props.item.MediaType=='SATA'">
                        <v-select
                          :items="IopsClassList_SATA"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.IopsClass"
                          dark></v-select>
                      </td>
                      <td v-if="props.item.MediaType=='SSD'">
                        <v-select
                          :items="IopsClassList_SSD"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.IopsClass"
                          dark></v-select>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Size"
                          :ref="'Volumn_Size_' + props.item.index"
                          required
                          :rules="rules.Services[item.Id].Volumns[props.item.Id].Size"
                          @input="rules.Services[item.Id].Volumns[props.item.Id].Size = rules0.Services.Volumns.Size"
                          placeholder="0表示不限制大小"
                          :class="{ 'last-field': item.Volumns.length == (props.item.index + 1) }"
                          add_func="addVolumn"
                          :add_params="item.Id"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-btn icon class="red--text text--lighten-2" @click.native="removei(item.Volumns, props.item.index)" title="删除">
                          <v-icon>close</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index < item.Volumns.length - 1" icon class="blue--text blue--lighten-2 ml-2" @click.native="downward(item.Volumns, props.item.index)" title="下移">
                          <v-icon>arrow_downward</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index > 0" icon class="green--text green--lighten-2 ml-2" @click.native="upward(item.Volumns, props.item.index)" title="上移">
                          <v-icon>arrow_upward</v-icon>
                        </v-btn>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4>
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>标签</v-subheader>
                    <v-spacer></v-spacer>
                    <v-btn icon class="blue--text text--lighten-2" @click.native="addLabel(item)">
                      <v-icon light>add</v-icon>
                    </v-btn>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_labels"
                    :items="item.Labels"
                    :customSort="nosort"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.Name"
                          :ref="'Label_Name_' + props.item.index"
                          required
                          :rules="rules.Services[item.Id].Labels[props.item.Id].Name"
                          @input="rules.Services[item.Id].Labels[props.item.Id].Name = rules0.Services.Labels.Name"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Value"
                          :ref="'Label_Value_' + props.item.index"
                          required
                          :class="{ 'completer-field' : true, 'last-field': item.Labels.length == (props.item.index + 1) }"
                          :rel="'Label_Value_' + props.item.index"
                          add_func="addLabel"
                          :add_params="item.Id"
                        ></v-text-field>
                      </td>
                      <td>
                        <v-btn icon class="red--text text--lighten-2" @click.native="removei(item.Labels, props.item.index)" title="删除">
                          <v-icon>close</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index < item.Labels.length - 1" icon class="blue--text blue--lighten-2 ml-2" @click.native="downward(item.Labels, props.item.index)" title="下移">
                          <v-icon>arrow_downward</v-icon>
                        </v-btn>
                        <v-btn v-if="props.item.index > 0" icon class="green--text green--lighten-2 ml-2" @click.native="upward(item.Labels, props.item.index)" title="上移">
                          <v-icon>arrow_upward</v-icon>
                        </v-btn>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
              </v-layout>
            </v-container>
          </div>
        </v-card>
        <div style="color:#9F9F9F;">
          提示：环境变量及标签中的值可以引用参数目录中的参数名，例如：一个表示域名的环境变量可以定义为“eureka1.${DOMAIN_SUFFIX}”。
        </div>
      </div>
    </v-flex>
    <v-flex xs12>
      <v-alert 
            v-if="alertArea==='CreateTemplate'"
            v-bind:success="alertType==='success'" 
            v-bind:info="alertType==='info'" 
            v-bind:warning="alertType==='warning'" 
            v-bind:error="alertType==='error'" 
            v-model="alertDisplay" 
            dismissible>{{ alertMsg }}</v-alert>
    </v-flex>
    <v-flex xs12 class="text-xs-center" mt-4>
      <v-btn class="orange darken-2 white--text" @click.native="save">
        <v-icon light left>save</v-icon>保存
      </v-btn>   
    </v-flex>
  </v-layout>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from '../api/api'
  import jQuery from 'jquery'
  import caret from '../caret'
  import completer from '../completer'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        headers_envs: [
          { text: '变量名', sortable: false, left: true },
          { text: '变量值', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        headers_ports: [
          { text: '容器端口', sortable: false, left: true },
          { text: '负载均衡ID', sortable: false, left: true },
          { text: '负载均衡目标群组ARN', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        headers_volumns: [
          { text: '容器挂载路径', sortable: false, left: true },
          { text: '卷类型', sortable: false, left: true },
          { text: '磁盘介质', sortable: false, left: true },
          { text: '读写频率', sortable: false, left: true },
          { text: '卷大小 (MB)', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        headers_labels: [
          { text: '标签名', sortable: false, left: true },
          { text: '标签值', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],

        MountTypeList: [
          { 'Label': '宿主机目录', Value: 'Directory' },
          { 'Label': '独占磁盘', Value: 'Disk' }
        ],

        MediaTypeList: [
          { 'Label': 'SATA', Value: 'SATA' },
          { 'Label': 'SSD', Value: 'SSD' }
        ],

        IopsClassList_SATA: [
          { 'Label': '很少', Value: 1 },
          { 'Label': '较少', Value: 2 },
          { 'Label': '中等', Value: 3 },
          { 'Label': '较重', Value: 4 },
          { 'Label': '很重', Value: 5 }
        ],

        IopsClassList_SSD: [
          { 'Label': '很少', Value: 6 },
          { 'Label': '较少', Value: 7 },
          { 'Label': '中等', Value: 8 },
          { 'Label': '较重', Value: 9 },
          { 'Label': '很重', Value: 10 }
        ],

        svcIdStart: 0,
        envIdStart: 0,
        portIdStart: 0,
        volumnIdStart: 0,
        labelIdStart: 0,

        Id: null,
        Title: '',
        Name: '',
        Version: '',
        Description: '',
        Services: [],

        Importing: false,

        TemplateData: '',
        TemplateDataDlg: false,

        EnvList: '',
        EnvListDlg: false,
        CurrentService: null,

        ApplicationNameWarning: false,
        ServiceNameWarning: false,
        NetworkModeWarning: false,
        OriginalValue: '',

        rules: {
          Services: []
        },

        rules0: {
          Title: [
            v => (v && v.length > 0 ? true : '请输入应用名称')
          ],
          Name: [
            v => (v && v.length > 0 ? (v.match(/\s/) ? "应用ID不允许包含空格" : (/^[a-z]+[a-z0-9\-]*$/.test(v) ? true : '应用ID只能由小写英文字母、减号、数字组成，并且以英文字母开头')) : '请输入应用ID')
          ],
          Version: [
            v => (v && v.length > 0 ? true : '请输入应用版本号')
          ],
          Services: {
            Title: [
              v => (v && v.length > 0 ? true : '请输入服务名称')
            ],
            Name: [
              v => (v && v.length > 0 ? (v.match(/\s/) ? "服务ID不允许包含空格" : (/^[a-z]+[a-z0-9\-]*$/.test(v) ? true : '服务ID只能由小写英文字母、减号、数字组成，并且以英文字母开头')) : '请输入应用ID')
            ],
            ImageName: [
              v => (v && v.length > 0 ? (v.match(/\s/) ? '镜像名称不允许包含空格' : true) : '请输入镜像名称')
            ],
            ImageTag: [
              v => (v && v.length > 0 ? (v.match(/\s/) ? '镜像Tag不允许包含空格' : true) : '请输入镜像Tag')
            ],
            CPU: [
              function(o) {
                let v = o ? o.toString() : '';
                return (v && v.length > 0 ? (/^\d+\.?\d*$/.test(v) && parseFloat(v) > 0 ? true : 'CPU个数必须大于0，可以为小数') : true)
              }
            ],
            Memory: [
              function(o) {
                let v = o ? o.toString() : '';
                return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 ? true : '内存必须为大于0的整数') : '请输入内存大小')
              }
            ],
            ReplicaCount: [
              function(o) {
                let v = o ? o.toString() : '';
                return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 && parseInt(v) <= 1000 ? true : '容器个数必须为1-1000的整数') : '请输入容器个数')
              }
            ],
            ServiceTimeout: [
              function(o) {
                  let v = o ? o.toString() : '';
                  return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 && parseInt(v) <= 1000 ? true : '服务启动等待时间必须为1-1000的整数') : '请输入服务启动等待时间')
                }
              ],
            Envs: { 
              Name: [
                v => (v && v.length > 0 ? (v.match(/\s/) ? '环境变量名称不允许包含空格' : true) : '请输入环境变量名称')
              ]
            },
            Ports: { 
              SourcePort: [
                function(o) {
                  let v = o ? o.toString() : '';
                  return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 && parseInt(v) <= 65535 ? true : '容器端口号必须为1-65535的整数') : '请输入容器端口号')
                }
              ]
            },
            Volumns: {
              ContainerPath: [
                v => (v && v.length > 0 ? (v.match(/\s/) ? '数据卷挂载路径不允许包含空格' : true) : '请输入数据卷挂载路径')
              ],
              Size: [
                function(o) {
                  let v = o ? o.toString() : '';
                  return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) >= 0 && parseInt(v) <= 4000000 ? true : '卷大小必须为0-4000000的整数') : '请输入卷大小')
                }
              ]
            },
            Labels: {
              Name: [
                v => (v && v.length > 0 ? (v.match(/\s/) ? '标签名称不允许包含空格' : true) : '请输入标签名称')
              ]
            }
          }
        }
      }
    },

    computed: {
      ...mapGetters([
          'alertArea',
          'alertType',
          'alertMsg'
      ]),

      alertDisplay: {
        get() {
          return this.$store.getters.alertArea != null;
        },
        set(v) {
          this.$store.dispatch('alertArea', null);
        }
      }
    },

    // 如果用router.replace做跳转，则需watch route，并且重新获取params中的参数
    watch: {
      '$route': 'init'
    },

    mounted() {
      this.init();
    },

    destroyed() {
      ui.showAlertAt();
    },

    methods: {
      init() {
        ui.showAlertAt('CreateTemplate');

        this.Id = this.$route.params ? this.$route.params.id : null;
        this.Title = this.$route.params ? this.$route.params.title : '';
        if (!this.Id) {
          return;
        }

        api.Template(this.Id).then(data => {
          this.initWithTemplateData(data);
        })
      },

      initWithTemplateData(data) {
        this.svcIdStart = 0;
        this.envIdStart = 0;
        this.portIdStart = 0;
        this.volumnIdStart = 0;
        this.labelIdStart = 0;

        this.Id = data.Id;
        this.Title = data.Title;
        this.Name = data.Name;
        this.Version = data.Version;
        this.Description = data.Description;

        let rules = {
          Title: this.rules0.Title,
          Name: this.rules0.Name,
          Version: this.rules0.Version,
          Services: []
        };

        if (!data.Services) {
          data.Services = [];
        } else {
          for (let st of data.Services) {
            st.index = st.Id = this.svcIdStart++;
            st.hidden = true;
            st.NetworkModeWarning = false;

            let r = {
              Title: this.rules0.Services.Title,
              Name: this.rules0.Services.Name,
              ImageName: this.rules0.Services.ImageName,
              ImageTag: this.rules0.Services.ImageTag,
              CPU: this.rules0.Services.CPU,
              Memory: this.rules0.Services.Memory,
              ReplicaCount: this.rules0.Services.ReplicaCount,
              ServiceTimeout: this.rules0.Services.ServiceTimeout,
              Envs: [],
              Ports: [],
              Volumns: [],
              Labels: []
            };

            if (!st.Envs) {
              st.Envs = [];
            } else {
              let i = 0;
              for (let e of st.Envs) {
                e.index = i++;
                e.Id = this.envIdStart++;
                r.Envs[e.Id] = this.rules0.Services.Envs;
              }
            }

            if (!st.Ports) {
              st.Ports = [];
            } else {
              let i = 0;
              for (let e of st.Ports) {
                e.index = i++;
                e.Id = this.portIdStart++;
                r.Ports[e.Id] = this.rules0.Services.Ports;
              }
            }

            if (!st.Volumns) {
              st.Volumns = [];
            } else {
              let i = 0;
              for (let e of st.Volumns) {
                e.index = i++;
                e.Id = this.volumnIdStart++;
                r.Volumns[e.Id] = this.rules0.Services.Volumns;
              }
            }

            if (!st.Labels) {
              st.Labels = [];
            } else {
              let i = 0;
              for (let e of st.Labels) {
                e.index = i++;
                e.Id = this.labelIdStart++;
                r.Labels[e.Id] = this.rules0.Services.Labels;
              }
            }

            rules.Services[st.Id] = r; 
          }
        }

        this.rules = rules;
        this.Services = data.Services;

        if (this.$route.params && this.$route.params.title) {
          this.Id = null;
          this.Title = this.$route.params.title;
        }

        this.initCompleters(); 
      },

      goback() {
        this.$router.go(-1);
      },

      initCompleters() {
        this.$nextTick(function() {
            let that = this;
            jQuery('.completer-field').each(function(e) {
              let input = jQuery(this).find('input');
              if (input.hasClass('with-completer')) {
                return;
              }

              input.addClass('with-completer');
              input.completer({
                url: that.$axios.defaults.baseURL + '/envs/values/search',
                completeSuggestion: function(e, v) {
                  let rel = e.parents('.completer-field').attr('rel');
                  Object.keys(that.$refs).forEach(k => {
                    if (k != rel) {
                      return;
                    }

                    let r = that.$refs[k];
                    if (Array.isArray(r)) {
                      r = r[0];
                    }

                    r.value = v;
                    r.inputValue = v;
                  });
                }
              });
            });

            jQuery('.last-field').each(function() {
              let input = jQuery(this).find('input');
              if (input.hasClass('with-hotkey')) {
                return;
              }

              input.addClass('with-hotkey');
              input.keydown(function(e) {
                if (e.keyCode == 9) {
                  let j = jQuery(this).parents('.last-field');
                  let f = j.attr('add_func');
                  let s = that.Services[j.attr('add_params')];
                  that[f](s);
                }
              });
            });
          });
      },

      showApplicationNameWarning(v) {
        if (v != this.OriginalValue) {
          this.ApplicationNameWarning = true;
        }
      },

      showServiceNameWarning(v) {
        if (v != this.OriginalValue) {
          this.ServiceNameWarning = true;
        }
      },

      mediaTypeChanged(item) {
        if (item.MediaType == 'SATA') {
          item.IopsClass = 3;
        } else if (item.MediaType == 'SSD') {
          item.IopsClass = 8;
        }
      },

      addService() {
        let id = this.svcIdStart++;
        this.$set(this.rules.Services, id, {});
        this.Services.push({
          Id: id,
          Title: '',
          Name: '',
          ImageName: '',
          ImageTag: '',
          CPU: '',
          ExclusiveCPU: false,
          Memory: '',
          ReplicaCount: '',
          ServiceTimeout: '10',
          NetworkMode: 'bridge',
          Description: '',
          Command: '',
          Restart: 'always',
          Envs: [],
          Ports: [],
          Volumns: [],
          Labels: [],
          hidden: false
        });
      },

      addEnv(s, e) {
        let id = this.envIdStart++;
        if (!this.rules.Services[s.Id].Envs) {
          this.rules.Services[s.Id].Envs = [];
        }

        this.$set(this.rules.Services[s.Id].Envs, id, {});

        if (!e) {
          e = { Name: '', Value: '' };
        }

        e.index = s.Envs.length;
        e.Id = id;

        s.Envs.push(e);
        this.patch(s.Envs);

        this.initCompleters();
      },

      addPort(s) {
        let id = this.portIdStart++;
        if (!this.rules.Services[s.Id].Ports) {
          this.rules.Services[s.Id].Ports = [];
        }

        this.$set(this.rules.Services[s.Id].Ports, id, {});
        
        s.Ports.push({ index: s.Ports.length, Id: id, SourcePort: '', LoadBalancerId: '', TargetGroupArn: '' });
        this.patch(s.Ports);

        this.initCompleters();
      },

      addVolumn(s) {
        let id = this.volumnIdStart++;
        if (!this.rules.Services[s.Id].Volumns) {
          this.rules.Services[s.Id].Volumns = [];
        }

        this.$set(this.rules.Services[s.Id].Volumns, id, {});
        
        s.Volumns.push({ index: s.Volumns.length, Id: id, ContainerPath: '', MountType: 'Directory', MediaType: 'SATA', IopsClass: 3, Size: 0 });
        this.patch(s.Volumns);

        this.initCompleters();
      },

      addLabel(s) {
        let id = this.labelIdStart++;
        if (!this.rules.Services[s.Id].Labels) {
          this.rules.Services[s.Id].Labels = [];
        }

        this.$set(this.rules.Services[s.Id].Labels, id, {});
        
        s.Labels.push({ index: s.Labels.length, Id: id, Name: '', Value: '' });
        this.patch(s.Labels);

        this.initCompleters();
      },

      downward(items, i) {
        let a = items[i];
        let b = items[i + 1];
        this.$set(items, i, b);
        this.$set(items, i + 1, a);
        this.patch(items);
      },

      upward(items, i) {
        let a = items[i];
        let b = items[i - 1];
        this.$set(items, i, b);
        this.$set(items, i - 1, a);
        this.patch(items);
      },

      removei(items, i) {
        items.splice(i, 1);
        this.patch(items);
      },

      hideService(item, h) {
        item.hidden = h;
      },

      /* Vuetify当前版本没有在slot中传递props.index，所以我们在item中预先设置index */
      patch(items) {
        let i = 0;
        for (let item of items) {
          item.index = i++;
        }
      },

      exportEnvs(item) {
        let s = "";
        for (let e of item.Envs) {
          if (s.length > 0) {
            s += "\n";
          }

          s += e.Name + "=" + e.Value;
        }

        this.EnvList = s;
        this.Importing = false;
        this.EnvListDlg = true;
      },

      importEnvs(item) {
        this.EnvList = '';
        this.Importing = true;
        this.CurrentService = item;
        this.EnvListDlg = true;
      },

      doImportEnvs() {
        let lines = this.EnvList.split(/\n/);
        for (let line of lines) {
          let p = line.indexOf('=');
          if (p > 0) {
            let n = line.substring(0, p);
            let v = line.substring(p + 1);
            this.addEnv(this.CurrentService, { Name: n, Value: v });
          }
        }

        this.EnvListDlg = false;
      },

      exportTemplate() {
        let t = {
          Title: this.Title,
          Name: this.Name,
          Version: this.Version,
          Description: this.Description,
          Services: this.Services
        };

        this.TemplateData = JSON.stringify(t, null, 4);
        this.Importing = false;
        this.TemplateDataDlg = true;
      },

      importTemplate() {
        this.TemplateData = '';
        this.Importing = true;
        this.TemplateDataDlg = true;
      },

      doImportTemplate() {
        let t;
        try {
          t = JSON.parse(this.TemplateData);
        } catch (e) {
          ui.alert('模板数据格式不正确');
          return;
        }

        this.initWithTemplateData(t);
        this.TemplateDataDlg = false;
      },

      save() {
        let rules = {
          Title: this.rules0.Title,
          Name: this.rules0.Name,
          Version: this.rules0.Version,
          Services: []
        };

        for (let t of this.Services) {
          let r = {
            Title: this.rules0.Services.Title,
            Name: this.rules0.Services.Name,
            ImageName: this.rules0.Services.ImageName,
            ImageTag: this.rules0.Services.ImageTag,
            CPU: this.rules0.Services.CPU,
            Memory: this.rules0.Services.Memory,
            ReplicaCount: this.rules0.Services.ReplicaCount,
            ServiceTimeout: this.rules0.Services.ServiceTimeout,
            Envs: [],
            Ports: [],
            Volumns: [],
            Labels: []
          };

          for (let e of t.Envs) {
            r.Envs[e.Id] = this.rules0.Services.Envs;
          }

          for (let e of t.Ports) {
            r.Ports[e.Id] = this.rules0.Services.Ports;
          }
          
          for (let e of t.Volumns) {
            r.Volumns[e.Id] = this.rules0.Services.Volumns;
          }

          for (let e of t.Labels) {
            r.Labels[e.Id] = this.rules0.Services.Labels;
          }

          rules.Services[t.Id] = r;
        }

        this.rules = rules;

        this.$nextTick(_ => {
          if (!this.validateForm()) {
            ui.alert('请正确填写应用模板');
            return;
          }

          for (let s of this.Services) {
            for (let e of s.Envs) {
              if (e.Value) {
                e.Value = e.Value.trim();
              }
            }

            for (let p of s.Ports) {
              if (p.LoadBalancerId) {
                p.LoadBalancerId = p.LoadBalancerId.trim();
              }

              if (p.TargetGroupArn) {
                p.TargetGroupArn = p.TargetGroupArn.trim();
              }
            }
          }

          let a = {
            Id: this.Id,
            Title: this.Title,
            Name: this.Name,
            Version: this.Version,
            Description: this.Description,
            Services: this.Services
          }

          if (this.Id && this.Id.length > 0) {
            api.UpdateTemplate(a).then(data => {
              ui.alert('应用模板修改成功', 'success');
              this.init();
            });
          } else {
            api.CreateTemplate(a).then(data => {
              ui.alert('新增应用模板成功', 'success');
              this.$router.replace('/templates/' + data.Id);
            });
          }
        });
      }
    }
  }
</script>

<style lang="stylus">
.completer-container
  font-family: inherit;
  font-size: 14px;
  line-height: normal;
  position: absolute;
  -webkit-box-sizing: border-box;
  -moz-box-sizing: border-box;
  box-sizing: border-box;
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid #ccc;
  border-bottom-color: #39f;
  background-color: #fff;

.completer-container li
  overflow: hidden;
  margin: 0;
  padding: .5em .8em;
  cursor: pointer;
  white-space: nowrap;
  text-overflow: ellipsis;
  border-bottom: 1px solid #eee;
  background-color: #fff;

.completer-container 
  .completer-selected, li:hover
    margin-left: -1px;
    border-left: 1px solid #39f;
    background-color: #eee;

.input-group--text-field
  &.env-list
    textarea
      font-size: 12px;
</style>
