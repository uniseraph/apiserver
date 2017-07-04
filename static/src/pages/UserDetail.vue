<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;用户列表&nbsp;&nbsp;/&nbsp;&nbsp;{{ Name }}
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
              single-line
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
              single-line
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
  import router from '../router'
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        Id: '',
        Name: '',
        Email: '',
        Tel: '',
        CreatedTime: 0,
        IsSysAdmin: false,
        IsAppAdmin: false,

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
          ]
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.User(this.$route.params.id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
          this.Email = data.Email;
          this.Tel = data.Tel;
          this.CreatedTime = data.CreatedTime;
          this.IsSysAdmin = (data.RoleSet & this.constants.ROLE_SYS_ADMIN) != 0;
          this.IsAppAdmin = (data.RoleSet & this.constants.ROLE_APP_ADMIN) != 0;
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

          let roleSet = this.constants.ROLE_NORMAL_USER;
          if (this.IsSysAdmin) {
            roleSet |= this.constants.ROLE_SYS_ADMIN;
          }
          if (this.IsAppAdmin) {
            roleSet |= this.constants.ROLE_APP_ADMIN;
          }

          api.UpdateUser({
            Id: this.Id,
            Name: this.Name,
            Email: this.Email,
            Tel: this.Tel,
            RoleSet: roleSet
          }).then(data => {
            ui.alert('用户资料修改成功', 'success');
            let that = this;
            setTimeout(() => {
              that.goback();
            }, 1500);
          });
        });
      }
    }
  }

</script>

<style lang="stylus">

</style>
