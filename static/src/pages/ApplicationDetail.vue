<template>
  <v-layout column>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
          &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;{{ PoolName }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ Title }}
          <v-spacer></v-spacer>
          <router-link :to="'/applications/' + Id + '/upgrade/' + encodeURIComponent(PoolName)" style="text-decoration:none;">
            <v-btn class="green darken-2 white--text" small>
              <v-icon light left>open_in_browser</v-icon>升级
            </v-btn>
          </router-link>
          <router-link :to="'/applications/' + Id + '/rollback/' + encodeURIComponent(PoolName)" style="text-decoration:none;">
            <v-btn class="orange darken-2 white--text ml-4" small>
              <v-icon light left>replay</v-icon>回滚
            </v-btn>
          </router-link>
        </v-card-title>
        <div>
          <v-container fluid>
            <v-layout row wrap>
              <v-flex xs2>
                <v-subheader>应用名称</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Title"
                  readonly
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>应用ID</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Name"
                  readonly
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
                <v-subheader>应用版本</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Version"
                  readonly
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
                  readonly
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
      </v-card-title>
      <div>
        <v-card v-for="(item, index) in Services" :key="item.Id" class="mb-2">
          <v-card-title>
            服务{{ index + 1 }}: {{ item.Title }}&nbsp;&nbsp;&nbsp;&nbsp;
            <span style="color:#9F9F9F;">
              域名: {{ Name }}-{{ item.Name }}.${DOMAIN_SUFFIX}
            </span>&nbsp;&nbsp;&nbsp;&nbsp;
            [&nbsp;<router-link :to="'/applications/containers/' + Id + '/' + item.Name + '/' + encodeURIComponent(PoolName) + '/' + encodeURIComponent(Title) + '/' + encodeURIComponent(item.Title)" style="text-decoration:none;">容器列表</router-link>&nbsp;]
            <v-spacer></v-spacer>
            <v-btn v-if="item.hidden" outline small icon class="blue blue--text mr-2" @click.native="hideService(item, false)" title="展开">
              <v-icon>arrow_drop_down</v-icon>
            </v-btn>
            <v-btn v-if="!item.hidden" outline small icon class="blue blue--text mr-2" @click.native="hideService(item, true)" title="折叠">
              <v-icon>arrow_drop_up</v-icon>
            </v-btn>
          </v-card-title>
          <div>
            <v-alert 
                  v-if="alertArea==='Service_' + item.Id"
                  v-bind:success="alertType==='success'" 
                  v-bind:info="alertType==='info'" 
                  v-bind:warning="alertType==='warning'" 
                  v-bind:error="alertType==='error'" 
                  v-model="alertMsg" 
                  dismissible>{{ alertMsg }}</v-alert>
          </div>
          <div v-show="!item.hidden">
            <v-container fluid>
              <v-layout row wrap>
                <v-flex xs2>
                  <v-subheader>服务名称</v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    v-model="item.Title"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>服务ID</v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    v-model="item.Name"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>镜像名称</v-subheader>
                </v-flex>
                <v-flex xs5>
                  <v-text-field
                    v-model="item.ImageName"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>镜像Tag</v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    v-model="item.ImageTag"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>CPU个数</v-subheader>
                </v-flex>
                <v-flex xs1>
                  <v-text-field
                    v-model="item.CPU"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-checkbox label="独占" v-model="item.ExclusiveCPU" dark disabled></v-checkbox>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>内存 (MB)</v-subheader>
                </v-flex>
                <v-flex xs3>
                  <v-text-field
                    v-model="item.Memory"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>容器个数<span class="required-star">*</span></v-subheader>
                </v-flex>
                <v-flex xs1>
                  <v-text-field
                    :ref="'Service_ReplicaCount_' + item.Id"
                    v-model="item.ReplicaCount"
                    required
                    :rules="rules.Services[item.Id].ReplicaCount"
                    @input="rules.Services[item.Id].ReplicaCount = rules0.Services.ReplicaCount"
                  ></v-text-field>
                </v-flex>
                <v-flex xs2 v-if="!Scaling">
                  <v-btn outline small class="green--text green--lighten-2" @click.native="updateReplicaCount(item)">
                      修改
                  </v-btn>
                </v-flex>
                <v-flex xs2 v-if="Scaling">
                  <v-progress-linear v-bind:indeterminate="true"></v-progress-linear>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs5>
                  <v-checkbox label="使用宿主机网络" v-model="item.UseHostNetwork" dark disabled></v-checkbox>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>说明</v-subheader>
                </v-flex>
                <v-flex xs10>
                  <v-text-field
                    v-model="item.Description"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                  <v-subheader>启动命令</v-subheader>
                </v-flex>
                <v-flex xs10>
                  <v-text-field
                    v-model="item.Command"
                    readonly
                  ></v-text-field>
                </v-flex>
                <v-flex xs2>
                </v-flex>
                <v-flex xs3>
                  <v-checkbox label="异常终止后自动重启" v-model="item.Restart" true-value="always" false-value="no" dark disabled></v-checkbox>
                </v-flex>
                <v-flex xs7>
                </v-flex>
                <v-flex xs12 mt-5 v-if="item.Envs && item.Envs.length > 0">
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>环境变量</v-subheader>
                    <v-spacer></v-spacer>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_envs"
                    :items="item.Envs"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.Name"
                          readonly
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Value"
                          readonly
                        ></v-text-field>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4 v-if="item.Ports && item.Ports.length > 0">
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>端口声明</v-subheader>
                    <v-spacer></v-spacer>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_ports"
                    :items="item.Ports"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.SourcePort"
                          readonly
                        ></v-text-field>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.TargetGroupArn"
                          readonly
                        ></v-text-field>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4 v-if="item.Volumns && item.Volumns.length > 0">
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>数据卷</v-subheader>
                    <v-spacer></v-spacer>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_volumns"
                    :items="item.Volumns"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.ContainerPath"
                          readonly
                        ></v-text-field>
                      </td>
                      <td>
                        <v-select
                          :items="MountTypeList"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.MountType"
                          dark
                          disabled></v-select>
                      </td>
                      <td>
                        <v-select
                          :items="MediaTypeList"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.MediaType"
                          dark
                          disabled></v-select>
                      </td>
                      <td v-if="props.item.MediaType=='SATA'">
                        <v-select
                          :items="IopsClassList_SATA"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.IopsClass"
                          dark
                          disabled></v-select>
                      </td>
                      <td v-if="props.item.MediaType=='SSD'">
                        <v-select
                          :items="IopsClassList_SSD"
                          item-text="Label"
                          item-value="Value"
                          v-model="props.item.IopsClass"
                          dark
                          disabled></v-select>
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Size"
                          readonly
                        ></v-text-field>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
                <v-flex xs12 mt-4 v-if="item.Labels && item.Labels.length > 0">
                  <v-divider></v-divider>
                  <v-card-title>
                    <v-subheader>标签</v-subheader>
                    <v-spacer></v-spacer>
                  </v-card-title>
                  <v-data-table
                    :headers="headers_labels"
                    :items="item.Labels"
                    hide-actions
                    class="elevation-1"
                    no-data-text=""
                  >
                    <template slot="items" scope="props">
                      <td>
                        <v-text-field
                          v-model="props.item.Name"
                      </td>
                      <td>
                        <v-text-field
                          v-model="props.item.Value"
                          readonly
                        ></v-text-field>
                      </td>
                    </template>
                  </v-data-table>
                </v-flex>
              </v-layout>
            </v-container>
          </div>
        </v-card>
      </div>
    </v-flex>
    <v-flex xs12>
      <v-card-title style="padding-left:0;">
        &nbsp;&nbsp;授权应用查看与SSH登录
        <v-spacer></v-spacer>
      </v-card-title>
      <div>
        <v-alert 
              v-if="alertArea==='Authorization'"
              v-bind:success="alertType==='success'" 
              v-bind:info="alertType==='info'" 
              v-bind:warning="alertType==='warning'" 
              v-bind:error="alertType==='error'" 
              v-model="alertMsg" 
              dismissible>{{ alertMsg }}</v-alert>
      </div>
      <v-layout row wrap>
        <v-flex xs6>
          <v-card>
            <v-card-title>
              授权团队
              <v-spacer></v-spacer>
              <v-select
                  :items="UnauthorizedTeamList"
                  label="请选择"
                  item-text="Name"
                  item-value="Id"
                  v-model="AuthorizeToTeam"
                  dark
                  max-height="auto"
                  autocomplete
                >
              </v-select>
              <v-btn floating small primary @click.native="addTeam">
                <v-icon light>add</v-icon>
              </v-btn>
            </v-card-title>
            <div class="auth-teams">
              <v-data-table
                :items="AuthorizedTeamList"
                hide-actions
                class="elevation-1"
                no-data-text=""
              >
                <template slot="items" scope="props">
                  <td>{{ props.item.Name }}</td>
                  <td align="right">
                    <v-btn outline small class="orange orange--text" @click.native="removeTeam(props.item)">
                      <v-icon class="orange--text">close</v-icon>删除
                    </v-btn>
                  </td>
                </template>
              </v-data-table>
            </div>
          </v-card>
        </v-flex>
        <v-flex xs6>
          <v-card>
            <v-card-title>
              授权用户
              <v-spacer></v-spacer>
              <v-select
                  :items="UnauthorizedUserList"
                  label="请选择"
                  item-text="Name"
                  item-value="Id"
                  v-model="AuthorizeToUser"
                  dark
                  max-height="auto"
                  autocomplete
                >
              </v-select>
              <v-btn floating small primary @click.native="addUser">
                <v-icon light>add</v-icon>
              </v-btn>
            </v-card-title>
            <div class="auth-users">
              <v-data-table
                :items="AuthorizedUserList"
                hide-actions
                class="elevation-1"
                no-data-text=""
              >
                <template slot="items" scope="props">
                  <td>{{ props.item.Name }}</td>
                  <td align="right">
                    <v-btn outline small class="orange orange--text" @click.native="removeUser(props.item)">
                      <v-icon class="orange--text">close</v-icon>删除
                    </v-btn>
                  </td>
                </template>
              </v-data-table>
            </div>
          </v-card>
        </v-flex>
      </v-layout>
    </v-flex>
  </v-layout>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        headers_envs: [
          { text: '变量名', sortable: false, left: true },
          { text: '变量值', sortable: false, left: true }
        ],
        headers_ports: [
          { text: '容器端口', sortable: false, left: true },
          { text: '负载均衡目标群组ARN', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        headers_volumns: [
          { text: '容器挂载路径', sortable: false, left: true },
          { text: '卷类型', sortable: false, left: true },
          { text: '磁盘介质', sortable: false, left: true },
          { text: '读写频率', sortable: false, left: true },
          { text: '卷大小 (MB)', sortable: false, left: true }
        ],
        headers_labels: [
          { text: '标签名', sortable: false, left: true },
          { text: '标签值', sortable: false, left: true }
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
        volumnIdStart: 0,
        labelIdStart: 0,

        Id: this.$route.params.id,
        PoolName: this.$route.params.poolName,
        Title: '',
        Name: '',
        Version: '',
        Description: '',
        Services: [],

        AuthorizedTeamList: [],
        AuthorizedUserList: [],
        UnauthorizedTeamList: [],
        UnauthorizedUserList: [],
        AuthorizeToTeam: null,
        AuthorizeToUser: null,

        Scaling: false,

        rules: {
          Services: []
        },

        rules0: {
          Services: {
            ReplicaCount: [
              function(o) {
                let v = o ? o.toString() : '';
                return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) >= 0 && parseInt(v) <= 1000 ? true : '容器个数必须为0-1000的整数') : '请输入容器个数')
              }
            ]
          }
        }
      }
    },

    computed: {
      ...mapGetters([
          'alertArea',
          'alertType',
          'alertMsg'
      ])
    },

    mounted() {
      this.init();
    },

    destroyed() {
      ui.showAlertAt();
    },

    methods: {
      init() {
        api.Application(this.Id).then(data => {
          this.svcIdStart = 0;
          this.envIdStart = 0;
          this.volumnIdStart = 0;
          this.labelIdStart = 0;

          this.Id = data.Application.Id;
          this.Title = data.Application.Title;
          this.Name = data.Application.Name;
          this.Version = data.Application.Version;
          this.Description = data.Application.Description;

          let rules = {
            Title: this.rules0.Title,
            Name: this.rules0.Name,
            Version: this.rules0.Version,
            Services: []
          };

          let services = data.Application.Services;
          if (!services) {
            services = [];
          } else {
            for (let st of services) {
              st.index = st.Id = this.svcIdStart++;
              st.hidden = true;

              let r = {
                ReplicaCount: this.rules0.Services.ReplicaCount
              };

              rules.Services[st.Id] = r; 
            }
          }

          this.rules = rules;
          this.Services = services;

          this.AuthorizedTeamList = data.Teams ? data.Teams : [];
          this.AuthorizedUserList = data.Users ? data.Users : [];
          this.AuthorizeToTeam = null;
          this.AuthorizeToUser = null;

          api.Teams().then(data => {
            this.UnauthorizedTeamList = filterArray(data, this.AuthorizedTeamList, 'Id');
          })

          api.Users().then(data => {
            this.UnauthorizedUserList = filterArray(data, this.AuthorizedUserList, 'Id') 
          })
        })
      },

      goback() {
        this.$router.go(-1);
      },

      hideService(item, h) {
        item.hidden = h;
      },

      updateReplicaCount(s) {
        ui.showAlertAt('Service_' + s.Id);
        this.Scaling = true;
        api.ScaleService({
          Id: this.Id,
          ServiceName: s.Name,
          ReplicaCount: s.ReplicaCount
        }).then(data => {
          ui.alert('容器个数更新成功', 'success');
          this.Scaling = false;
        }, err => {
          this.Scaling = false;
        })
      },

      /* Vuetify当前版本没有在slot中传递props.index，所以我们在item中预先设置index */
      patch(items) {
        let i = 0;
        for (let item of items) {
          item.index = i++;
        }
      },

      addTeam() {
        if (this.AuthorizeToTeam) {
          ui.showAlertAt('Authorization');
          api.AddTeamToApplication({ Id: this.Id, TeamId: this.AuthorizeToTeam }).then(data => {
            this.init();
          })
        }
      },

      removeTeam(team) {
        ui.showAlertAt('Authorization');
        api.RemoveTeamFromApplication({ Id: this.Id, TeamId: team.Id }).then(data => {
            this.init();
          })
      },

      addUser() {
        if (this.AuthorizeToUser) {
          ui.showAlertAt('Authorization');
          api.AddUserToApplication({ Id: this.Id, UserId: this.AuthorizeToUser }).then(data => {
            this.init();
          })
        }
      },

      removeUser(user) {
        ui.showAlertAt('Authorization');
        api.RemoveUserFromApplication({ Id: this.Id, UserId: user.Id }).then(data => {
            this.init();
          })
      }
    }
  }

  function filterArray(arr1, arr2, p) {
    let m = array2Map(arr2, p);
    let r = [];
    for (let e of arr1) {
      if (!m.has(e[p])) {
        r.push(e);
      }
    }

    return r;
  }

  function array2Map(arr, p) {
    let m = new Map();
    for (let e of arr) {
      m.set(e[p], e);
    }

    return m;
  }
</script>

<style lang="stylus">

</style>
