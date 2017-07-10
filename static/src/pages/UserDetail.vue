<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;用户列表&nbsp;&nbsp;/&nbsp;&nbsp;{{ Id ? Name : '新增用户' }}
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
              required
              :rules="rules.Name"
              @input="rules.Name = rules0.Name"
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
              required
              :rules="rules.Email"
              @input="rules.Email = rules0.Email"
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
            <v-subheader>电话<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              ref="Tel"
              v-model="Tel"
              required
              :rules="rules.Tel"
              @input="rules.Tel = rules0.Tel"
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs2>
            <v-subheader>注册时间</v-subheader>
          </v-flex>
          <v-flex xs3>
            <div class="input-group--span">
              {{ CreatedTime | formatDateTime }}
            </div>
          </v-flex>
          <v-flex xs2>
            <v-subheader>角色</v-subheader>
          </v-flex>
          <v-flex xs10>
            <v-checkbox label="系统管理员" v-model="IsSysAdmin" dark></v-checkbox>
            <v-checkbox label="应用管理员" v-model="IsAppAdmin" dark></v-checkbox>
          </v-flex>
          <v-flex xs2>
            <v-subheader>初始密码<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Password1"
              ref="Password1"
              type="password"
              single-line
              required
              :rules="rules.Password1"
              @input="rules.Password1 = rules0.Password1"
            ></v-text-field>
          </v-flex>
          <v-flex xs7>
          </v-flex>
          <v-flex xs2>
            <v-subheader>再输一次<span class="required-star">*</span></v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Password2"
              ref="Password2"
              type="password"
              single-line
              required
              :rules="rules.Password2"
              @input="rules.Password2 = rules0.Password2"
            ></v-text-field>
          </v-flex>
          <v-flex xs7>
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
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Id: this.$route.params ? this.$route.params.id : null,
        Name: '',
        Email: '',
        Tel: '',
        CreatedTime: 0,
        IsSysAdmin: false,
        IsAppAdmin: false,
        Password1: '',
        Password2: '',

        rules: {},

        rules0: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入用户名')
          ],
          Email: [
            v => (v && v.length > 0 ? (this.isEmail(v) ? true : '邮箱格式不正确') : '请输入邮箱')
          ],
          Tel: [
            v => (v && v.length > 0 ? true : '请输入电话')
          ],
          Password1: [
            () => (this.Password1.length < 8 ? '密码至少需8位字符' : true)
          ],
          Password2: [
            () => (this.Password1 != this.Password2 ? '两次输入的密码不相同' : true)
          ]
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        if (this.Id) {
          api.User(this.Id).then(data => {
            this.Id = data.Id;
            this.Name = data.Name;
            this.Email = data.Email;
            this.Tel = data.Tel;
            this.CreatedTime = data.CreatedTime;
            this.IsSysAdmin = (data.RoleSet & this.constants.ROLE_SYS_ADMIN) != 0;
            this.IsAppAdmin = (data.RoleSet & this.constants.ROLE_APP_ADMIN) != 0;
          })
        }
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

          let roleSet = this.constants.ROLE_NORMAL_USER;
          if (this.IsSysAdmin) {
            roleSet |= this.constants.ROLE_SYS_ADMIN;
          }
          if (this.IsAppAdmin) {
            roleSet |= this.constants.ROLE_APP_ADMIN;
          }

          if (this.Id) {
            api.UpdateUser({
              Id: this.Id,
              Name: this.Name,
              Email: this.Email,
              Tel: this.Tel,
              RoleSet: roleSet
            }).then(data => {
              ui.alert('用户资料修改成功', 'success');
              this.init();
            });
          } else {
            api.CreateUser({
              Name: this.Name,
              Email: this.Email,
              Tel: this.Tel,
              RoleSet: roleSet,
              Pass: this.Password1
            }).then(data => {
              ui.alert('新增用户成功', 'success');
              this.$router.replace('/users/' + data.Id);
            });
          }
        });
      }
    }
  }

</script>

<style lang="stylus">

</style>
