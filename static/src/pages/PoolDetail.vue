<template>
  <v-layout column>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
          &nbsp;&nbsp;集群列表&nbsp;&nbsp;/&nbsp;&nbsp;{{ Name }}
          <v-spacer></v-spacer>
        </v-card-title>
        <div>
          <v-container fluid>
            <v-layout row wrap>
              <v-flex xs2>
                <v-subheader>名称<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Name"
                  ref="all_Name"
                  required
                  :rules="rules.Name"
                  @input="rules.Name = rules0.Name"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>参数目录</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="EnvTree.Name"
                  readonly
                ></v-text-field>
                <!--v-select
                  :items="EnvTreeList"
                  item-text="Name"
                  item-value="Id"
                  v-model="EnvTreeId"
                  label="请选择"
                  dark
                ></v-select-->
              </v-flex>
              <v-flex xs2>
                <v-subheader>驱动类型</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Driver"
                  readonly
                ></v-text-field>
                <!--v-select
                  :items="DriverList"
                  v-model="Driver"
                  ref="all_Driver"
                  label="请选择"
                  dark
                  required
                  :rules="rules.Driver"
                ></v-select-->
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>驱动版本</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-text-field
                  v-model="DriverOpts.Version"
                  readonly
                ></v-text-field>
                <!--v-select
                  :items="SwarmVersionList"
                  v-model="DriverOpts.Version"
                  ref="swarm_Version"
                  label="请选择"
                  dark
                  required
                  :rules="rules.DriverOpts.swarm.Version"
                ></v-select-->
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>API地址</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-text-field
                  v-model="DriverOpts.EndPoint"
                  readonly
                ></v-text-field>
                <!--v-text-field
                  v-model="DriverOpts.EndPoint"
                  ref="swarm_EndPoint"
                  required
                  :rules="rules.DriverOpts.swarm.EndPoint"
                  @input="rules.DriverOpts.swarm.EndPoint = rules0.DriverOpts.swarm.EndPoint"
                ></v-text-field-->
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>API版本</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-text-field
                  v-model="DriverOpts.APIVersion"
                  readonly
                ></v-text-field>
                <!--v-select
                  :items="SwarmAPIVersionList"
                  v-model="DriverOpts.APIVersion"
                  ref="swarm_APIVersion"
                  label="请选择"
                  dark
                  required
                  :rules="rules.DriverOpts.swarm.APIVersion"
                ></v-select-->
              </v-flex>
              <v-flex xs3>
                <v-subheader>节点个数：{{ Nodes }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>CPU：{{ CPUs }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>内存 (GB)：{{ Memory }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>磁盘 (GB)：{{ Disk }}</v-subheader>
              </v-flex>
              <v-flex xs12 mt-4 class="text-xs-center">
                <v-btn class="orange darken-2 white--text" @click.native="save">
                  <v-icon light left>save</v-icon>保存
                </v-btn>            
              </v-flex>
            </v-layout>
          </v-container>
        </div>
      </v-card>
    </v-flex>
    <v-flex xs12 mt-4>
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
                    <v-btn class="orange darken-2 white--text" small @click.native="removeTeam(props.item)">
                      <v-icon light left>close</v-icon>删除
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
              <v-btn floating small primary @click.native="addTeam">
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
                    <v-btn class="orange darken-2 white--text" small @click.native="removeUser(props.item)">
                      <v-icon light left>close</v-icon>删除
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
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Id: this.$route.params.id,
        Name: '',
        EnvTreeId: null,
        EnvTree: {},
        Driver: 'swarm',
        DriverOpts: { Version: 'v1.0', EndPoint: '', APIVersion: 'v1.23' },
        Nodes: 0,
        CPUs: 0,
        Memory: 0,
        Disk: 0,

        EnvTreeList: [],
        DriverList: [ 'swarm' ],
        SwarmVersionList: [ 'v1.0' ],
        SwarmAPIVersionList: [ 'v1.23' ],

        AuthorizedTeamList: [],
        AuthorizedUserList: [],
        UnauthorizedTeamList: [],
        UnauthorizedUserList: [],
        AuthorizeToTeam: null,
        AuthorizeToUser: null,

        rules: { 
          DriverOpts: { swarm: {} } 
        },

        rules0: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入集群名称')
          ],
          Driver: [
            v => (v && v.length > 0 ? true : '请选择驱动类型')
          ],
          DriverOpts: {
            swarm: {
              Version: [
                v => (v && v.length > 0 ? true : '请选择集群驱动版本')
              ],
              EndPoint: [
                v => (v && v.length > 0 ? true : '请输入集群API地址')
              ],
              APIVersion: [
                v => (v && v.length > 0 ? true : '请选择集群API版本')
              ]
            }
          }
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Pool(this.Id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
          this.EnvTreeId = data.EnvTreeId;
          this.EnvTree = data.EnvTree;
          this.Driver = data.Driver;
          this.DriverOpts = data.DriverOpts;
          this.Nodes = data.Nodes;
          this.CPUs = data.CPUs;
          this.Memory = data.Memory;
          this.Disk = data.Disk;
          this.AuthorizedTeamList = data.Teams ? data.Teams : [];
          this.AuthorizedUserList = data.Users ? data.Users : [];
          this.AuthorizeToTeam = null;
          this.AuthorizeToUser = null;

          api.EnvTrees().then(data => {
            this.EnvTreeList = data;
          })

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

      save() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.validateForm('all_') || !this.validateForm(this.Driver + '_')) {
            return;
          }

          api.UpdatePool({
            Id: this.Id,
            Name: this.Name,
            EnvTreeId: this.EnvTreeId,
            Driver: this.Driver,
            DriverOpts: this.DriverOpts
          }).then(data => {
            ui.alert('集群资料修改成功', 'success');
          });
        });
      },

      addTeam() {
        if (this.AuthorizeToTeam) {
          api.AddTeamToPool({ Id: this.Id, TeamId: this.AuthorizeToTeam }).then(data => {
            this.init();
          })
        }
      },

      removeTeam(team) {
        api.RemoveUserFromPool({ Id: this.Id, TeamId: team.Id }).then(data => {
            this.init();
          })
      },

      addUser() {
        if (this.AuthorizeToUser) {
          api.AddUserToPool({ Id: this.Id, UserId: this.AuthorizeToUser }).then(data => {
            this.init();
          })
        }
      },

      removeUser(user) {
        api.RemoveUserFromPool({ Id: this.Id, UserId: user.Id }).then(data => {
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
.auth-teams
  thead
    display: none 

.auth-users
  thead
    display: none 
</style>
