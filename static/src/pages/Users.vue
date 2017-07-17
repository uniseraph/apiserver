<template>
  <v-card>
    <v-card-title>
      用户列表
      <v-spacer></v-spacer>
      <router-link :to="'/users/create'">
        <v-btn class="primary white--text"><v-icon light>add</v-icon>新增用户</v-btn>
      </router-link>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="RemoveConfirmDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>提示</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>你确认要删除用户{{ SelectedUser.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeUser">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-data-table
        :headers="headers"
        :items="items"
        hide-actions
        class="users-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td><router-link :to="'/users/' + props.item.Id">{{ props.item.Name }}</router-link></td>
          <td>{{ props.item.Email }}</td>
          <td>{{ props.item.Tel }}</td>
          <td>
            <div v-if="props.item.RoleSet & constants.ROLE_SYS_ADMIN">系统管理员</div>
            <div v-if="props.item.RoleSet & constants.ROLE_APP_ADMIN">应用管理员</div>
            <div v-if="props.item.RoleSet == 1">普通用户</div>
          </td>
          <td>{{ props.item.CreatedTime | formatDate }}</td>
          <td>
            <router-link :to="'/users/password/' + props.item.Id">
              <v-btn outline small icon class="green green--text" title="重置密码">
                <v-icon>lock</v-icon>
              </v-btn>
            </router-link>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除用户">
              <v-icon>close</v-icon>
            </v-btn>
          </td>
        </template>
      </v-data-table>
    </div>
  </v-card>
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        headers: [
          { text: '用户名', sortable: false, left: true },
          { text: '邮箱', sortable: false, left: true },
          { text: '电话', sortable: false, left: true },
          { text: '角色', sortable: false, left: true },
          { text: '注册时间', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],
        
        RemoveConfirmDlg: false,
        SelectedUser: {}
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Users().then(data => {
          this.items = data;
        })
      },

      confirmBeforeRemove(user) {
        this.SelectedUser = user;
        this.RemoveConfirmDlg = true;
      },

      removeUser() {
        this.RemoveConfirmDlg = false;
        api.RemoveUser(this.SelectedUser.Id).then(data => {
          this.init();
        })
      }
    }
  }
</script>

<style lang="stylus">
.users-table
  tr
    .btn
      visibility: hidden
  tr:hover
    .btn
      visibility: visible
</style>
