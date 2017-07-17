<template>
  <v-layout column>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
          &nbsp;&nbsp;团队管理&nbsp;&nbsp;/&nbsp;&nbsp;{{ Name }}
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
                  ref="Name"
                  required
                  :rules="rules.Name"
                  @input="rules.Name = rules0.Name"
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
      <v-card>
        <v-card-title>
          团队成员
          <v-spacer></v-spacer>
          <v-select
              :items="UsersNotInTeam"
              label="请选择"
              item-text="Name"
              item-value="Id"
              v-model="UserToJoin"
              dark
              max-height="auto"
              autocomplete
            >
          </v-select>
          <v-btn floating small primary @click.native="addUser">
            <v-icon light>add</v-icon>
          </v-btn>
        </v-card-title>
        <div>
          <v-data-table
            :headers="headers"
            :items="UsersInTeam"
            hide-actions
            class="elevation-1"
            no-data-text=""
          >
            <template slot="items" scope="props">
              <td>{{ props.item.Id }}</td>
              <td>{{ props.item.Name }}</td>
              <td>{{ props.item.Email }}</td>
              <td>{{ props.item.Tel }}</td>
              <td>
                <div v-if="props.item.RoleSet & constants.ROLE_SYS_ADMIN">系统管理员</div>
                <div v-if="props.item.RoleSet & constants.ROLE_APP_ADMIN">应用管理员</div>
                <div v-if="props.item.RoleSet == 1">普通用户</div>
              </td>
              <td>
                <v-radio label="" v-model="LeaderId" :value="props.item.Id" dark></v-radio>
              </td>
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
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Id: this.$route.params.id,
        Name: '',
        Description: '',
        LeaderId: false,
        headers: [
          { text: 'ID', sortable: false, left: true },
          { text: '用户名', sortable: false, left: true },
          { text: '邮箱', sortable: false, left: true },
          { text: '电话', sortable: false, left: true },
          { text: '角色', sortable: false, left: true },
          { text: '主管', sortable: false, left: true }
        ],
        UsersInTeam: [],
        UsersNotInTeam: [],
        UserToJoin: null,

        rules: {},

        rules0: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入团队名称')
          ]
        }
      }
    },

    watch: {
      LeaderId(newLeaderId, oldLeaderId) {
        if (oldLeaderId === false) {
          return;
        }

        api.AppointLeader({
          TeamId: this.Id,
          UserId: newLeaderId
        }).then(data => {
        })
        .catch(err => {
          ui.alert('设置主管发生错误')
        });
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Team(this.Id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
          this.Description = data.Description;
          this.LeaderId = data.Leader ? data.Leader.Id : null;
          this.UsersInTeam = data.Users;
          this.UserToJoin = null,

          api.Users().then(data => {
            this.UsersNotInTeam = filterArray(data, this.UsersInTeam, 'Id') 
          })
        })
      },

      goback() {
        this.$router.go(-1);
      },

      save() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.validateForm()) {
            return;
          }

          api.UpdateTeam({
            Id: this.Id,
            Name: this.Name,
            Description: this.Description
          }).then(data => {
            ui.alert('团队资料修改成功', 'success');
          });
        });
      },

      addUser() {
        if (this.UserToJoin) {
          api.JoinTeam({ TeamId: this.Id, UserId: this.UserToJoin }).then(data => {
            this.init();
          })
        }
      },

      removeUser(user) {
        api.QuitTeam({ TeamId: this.Id, UserId: user.Id }).then(data => {
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
