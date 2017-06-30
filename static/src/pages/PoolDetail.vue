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
                <v-subheader>驱动类型</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-select
                  v-bind:items="DriverList"
                  v-model="Driver"
                  label="请选择"
                  dark
                  single-line
                  auto
                  required
                ></v-select>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-subheader>网络类型</v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-select
                  v-bind:items="NetworkList"
                  v-model="Network"
                  label="请选择"
                  dark
                  single-line
                  auto
                  required
                ></v-select>
              </v-flex>
              <v-flex xs2>
                <v-subheader>API地址</v-subheader>
              </v-flex>
              <v-flex xs10>
                <v-text-field
                  v-model="EndPoint"
                  required
                  single-line
                ></v-text-field>
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
                  v-bind:items="UnauthorizedTeamList"
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
                v-bind:items="AuthorizedTeamList"
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
                  v-bind:items="UnauthorizedUserList"
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
                v-bind:items="AuthorizedUserList"
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
        Driver: '',
        Network: '',
        EndPoint: '',
        Nodes: 0,
        Cpus: 0,
        Memories: 0,
        Disks: 0,
        DriverList: [ 'Swarm', 'Kubernetes' ],
        NetworkList: [ 'Flannel', 'VxLAN' ],
        AuthorizedTeamList: [],
        AuthorizedUserList: [],
        UnauthorizedTeamList: [],
        UnauthorizedUserList: [],
        AuthorizeToTeam: null,
        AuthorizeToUser: null
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
          this.Network = data.Network;
          this.EndPoint = data.EndPoint;
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

      save() {
        api.UpdatePool({
          Id: this.Id,
          Name: this.Name,
          Driver: this.Driver,
          Network: this.Network
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
