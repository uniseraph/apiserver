<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;用户列表&nbsp;&nbsp;/&nbsp;&nbsp;新增用户
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs2>
            <v-subheader>用户名<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              ref="Name"
              v-model="Name"
              single-line
              :rules="rules.NameRules"
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs2>
            <v-subheader>邮箱<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              ref="Email"
              v-model="Email"
              single-line
              :rules="rules.EmailRules"
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
            <v-subheader>电话<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              ref="Tel"
              v-model="Tel"
              single-line
              :rules="rules.TelRules"
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs3>
          </v-flex>
          <v-flex xs2>
            <v-subheader>角色</v-subheader>
          </v-flex>
          <v-flex xs10>
            <v-checkbox label="系统管理员" v-model="IsSysAdmin" dark></v-checkbox>
            <v-checkbox label="应用管理员" v-model="IsAppAdmin" dark></v-checkbox>
          </v-flex>
          <v-flex xs12 mt-4 class="text-md-center">
            <v-btn class="orange darken-2 white--text" @click.native="save">
              <v-icon light left>save</v-icon>保存
            </v-btn>            
          </v-flex>
        </v-layout>
      </v-container>
    </div>
  </v-card>
</template>

<script>
  import router from '../router'
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Name: '',
        Email: '',
        Tel: '',
        IsSysAdmin: false,
        IsAppAdmin: false,

        rules: {
          NameRules: [
            v => (v.length > 0 ? true : '请输入用户名')
          ],
          EmailRules: [
            v => (v.length > 0 ? (this.isEmail(v) ? true : '邮箱格式不正确') : '请输入邮箱')
          ],
          TelRules: [
            v => (v.length > 0 ? true : '请输入电话')
          ]
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {

      },

      goback() {
        router.go(-1);
      },

      save() {
        for (let f in this.$refs) {
          let e = this.$refs[f];
          if (e.errorBucket && e.errorBucket.length > 0) {
            ui.alert('请正确填写用户资料');
            return;
          }
        }

        let roleSet = this.constants.ROLE_NORMAL_USER;
        if (this.IsSysAdmin) {
          roleSet |= this.constants.ROLE_SYS_ADMIN;
        }
        if (this.IsAppAdmin) {
          roleSet |= this.constants.ROLE_APP_ADMIN;
        }

        api.CreateUser({
          Name: this.Name,
          Email: this.Email,
          Tel: this.Tel,
          RoleSet: roleSet
        }).then(data => {
          ui.alert('新增用户成功', 'success');
          let that = this;
          setTimeout(() => {
            that.goback();
          }, 1500);
        })
      }
    }
  }

</script>

<style lang="stylus">

</style>
