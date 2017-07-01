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
                  single-line
                  :rules="rules.Name"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>驱动类型<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-select
                  :items="DriverList"
                  v-model="Driver"
                  ref="all_Driver"
                  label="请选择"
                  dark
                  single-line
                  :rules="rules.Driver"
                ></v-select>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>驱动版本</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-select
                  :items="SwarmVersionList"
                  v-model="DriverOpts.Version"
                  ref="swarm_Version"
                  label="请选择"
                  dark
                  single-line
                  :rules="rules.DriverOpts.swarm.Version"
                ></v-select>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>API地址</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-text-field
                  v-model="DriverOpts.EndPoint"
                  ref="swarm_EndPoint"
                  single-line
                  :rules="rules.DriverOpts.swarm.EndPoint"
                ></v-text-field>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs2>
                <v-subheader>驱动版本</v-subheader>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs3>
                <v-select
                  :items="SwarmAPIVersionList"
                  v-model="DriverOpts.APIVersion"
                  ref="swarm_APIVersion"
                  label="请选择"
                  dark
                  single-line
                  :rules="rules.DriverOpts.swarm.APIVersion"
                ></v-select>
              </v-flex>
              <v-flex v-if="Driver == 'swarm'" xs7>
              </v-flex>
              <v-flex xs3>
                <v-subheader>节点个数：{{ Nodes }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>CPU：{{ Cpus }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>内存 (GB)：{{ Memories }}</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-subheader>磁盘 (GB)：{{ Disks }}</v-subheader>
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
                  single-line
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
                  single-line
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
  import router from '../router'
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Id: '',
        Name: '',
        Driver: 'swarm',
        DriverOpts: { Version: 'v1.0', EndPoint: '', APIVersion: 'v1.23' },
        Nodes: 0,
        Cpus: 0,
        Memories: 0,
        Disks: 0,

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
        api.Pool(this.$route.params.id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
          this.Driver = data.Driver;
          this.DriverOpts = data.DriverOpts;
          this.Nodes = data.Nodes;
          this.Cpus = data.Cpus;
          this.Memories = data.Memories;
          this.Disks = data.Disks;
          this.AuthorizedTeamList = data.Teams;
          this.AuthorizedUserList = data.Users;
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
        router.go(-1);
      },

      validateForm(refPrefix) {
        for (let f in this.$refs) {
          if (f.indexOf(refPrefix) == 0) {
            let e = this.$refs[f];
            if (e.errorBucket && e.errorBucket.length > 0) {
              return false;
            }
          }
        }

        return true;
      },

      save() {
        if (!this.validateForm('all_') || !this.validateForm(this.Driver + '_')) {
          return;
        }

        api.UpdatePool({
          Id: this.Id,
          Name: this.Name,
          Driver: this.Driver,
          Network: this.Network,
          EndPoint: this.EndPoint
        }).then(data => {
          ui.alert('集群资料修改成功', 'success');
        })
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
