<template>
  <v-app light :class="{ 'login-ui': token && token.Id == null }">
    <div v-if="token && token.Id == null" class="text-xs-center login-content" primary>
      <v-layout row wrap class="login-box">
        <v-flex xs4 offset-xs4>
          <img src="/public/logo.png"></img>
        </v-flex>
        <v-flex xs4>
        </v-flex>
        <v-flex xs4 offset-xs4 class="mt-4">
          <v-card>
            <v-card-title class="text-xs-center" style="font-size:24px;display:block;">
              登录
            </v-card-title>
            <div>
              <v-container fluid>
                <v-layout row wrap>
                  <v-flex xs4>
                    <v-subheader>用户名</v-subheader>
                  </v-flex>
                  <v-flex xs8>
                    <v-text-field
                      v-model="Login.Name"
                      ref="Login_Name"
                      single-line
                      required
                      :rules="rules.Login.Name"
                      @input="rules.Login.Name = rules0.Login.Name"
                    ></v-text-field>
                  </v-flex>
                  <v-flex xs4>
                    <v-subheader>密码</v-subheader>
                  </v-flex>
                  <v-flex xs8>
                    <v-text-field
                      v-model="Login.Password"
                      ref="Login_Password"
                      single-line
                      required
                      type="password"
                      :rules="rules.Login.Password"
                      @input="rules.Login.Password = rules0.Login.Password"
                      @keydown.enter.native="login"
                    ></v-text-field>
                  </v-flex>
                  <v-flex xs4>
                  </v-flex>
                  <v-flex xs8 mt-4 class="text-xs-left">
                    <v-btn class="orange darken-2 white--text" @click.native="login">
                      <v-icon light left>vpn_key</v-icon>登录
                    </v-btn>     
                  </v-flex>
                </v-layout>
              </v-container>
            </div>
          </v-card>
        </v-flex>
        <v-flex xs4>
        </v-flex>
        <v-flex xs4 offset-xs4>
          <v-alert 
            v-if="alertArea==='global'"
            v-bind:success="alertType==='success'" 
            v-bind:info="alertType==='info'" 
            v-bind:warning="alertType==='warning'" 
            v-bind:error="alertType==='error'" 
            v-model="alertMsg" 
            dismissible>{{ alertMsg }}</v-alert>
        </v-flex>
        <v-flex xs4>
        </v-flex>
      </v-layout>
      <div class="login-footer">
      </div>
    </div>
    <div v-if="token && token.Id != null">
      <v-navigation-drawer
        persistent
        :mini-variant="miniVariant"
        v-model="drawer"
      >
        <v-list class="pa-0">
          <v-list-item>
            <v-list-tile avatar>
              <v-list-tile-avatar v-if="!miniVariant">
                <v-icon light>bubble_chart</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content v-if="!miniVariant">
                <v-list-tile-title>峥云网络</v-list-tile-title>
              </v-list-tile-content>
              <v-spacer></v-spacer v-if="!miniVariant">
              <v-list-tile-action>
                <v-btn light icon
                  @click.native.stop="miniVariant = !miniVariant">
                  <v-icon v-html="miniVariant ? 'chevron_right' : 'chevron_left'"></v-icon>
                </v-btn>
              </v-list-tile-action>
            </v-list-tile>
          </v-list-item>
        </v-list>
        <v-divider></v-divider>
        <v-subheader light>集群与应用管理</v-subheader>
        <v-list>
          <v-list-item>
            <v-list-tile ripple to="/dashboard" router>
              <v-list-tile-avatar>
                <v-icon light>insert_chart</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>信息总览</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/pools" router>
              <v-list-tile-avatar>
                <v-icon light>home</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>集群列表</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/env/trees" router>
              <v-list-tile-avatar>
                <v-icon light>local_parking</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>参数目录</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/templates" router>
              <v-list-tile-avatar>
                <v-icon light>share</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>应用模板</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/applications" router>
              <v-list-tile-avatar>
                <v-icon light>brightness_auto</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>应用管理</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/cicd" router>
              <v-list-tile-avatar>
                <v-icon light>cloud_upload</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>集成交付</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/logs" router>
              <v-list-tile-avatar>
                <v-icon light>library_books</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>系统日志</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/audit" router>
              <v-list-tile-avatar>
                <v-icon light>remove_from_queue</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>操作审计</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <!--v-list-item>
            <v-list-tile ripple to="/monitor" router>
              <v-list-tile-avatar>
                <v-icon light>report_problem</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>监控预警</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item-->
        </v-list>
        <v-subheader light>权限管理</v-subheader>
        <v-list>
          <v-list-item>
            <v-list-tile ripple to="/users" router>
              <v-list-tile-avatar>
                <v-icon light>group</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>用户列表</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
          <v-list-item>
            <v-list-tile ripple to="/teams" router>
              <v-list-tile-avatar>
                <v-icon light>group_add</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>团队管理</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list-item>
        </v-list>
      </v-navigation-drawer>
      <v-toolbar light>
        <v-toolbar-side-icon light @click.native.stop="drawer = !drawer">
        </v-toolbar-side-icon>
        <v-toolbar-title>峥云网络</v-toolbar-title>
        <v-spacer></v-spacer>
        <v-subheader light>{{ token.Name }}</v-subheader>
        <v-btn light icon @click.native="logout" title="退出">
          <v-icon>exit_to_app</v-icon>
        </v-btn>
      </v-toolbar>
      <main>
        <v-container fluid>
          <v-slide-y-transition mode="out-in">
            <v-layout column>
              <v-flex xs12>
                <v-alert 
                  v-if="alertArea==='global'"
                  v-bind:success="alertType==='success'" 
                  v-bind:info="alertType==='info'" 
                  v-bind:warning="alertType==='warning'" 
                  v-bind:error="alertType==='error'" 
                  v-model="alertMsg" 
                  dismissible>{{ alertMsg }}</v-alert>
              </v-flex>
              <v-flex xs12>
                <router-view></router-view>
              </v-flex>
            </v-layout>
          </v-slide-y-transition>
        </v-container>
      </main>
    </div>
  </v-app>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from './api/api'
  import * as ui from './util/ui'
  import * as context from './util/context'

  export default {
    data() {
      return {
        drawer: true,
        fixed: false,
        miniVariant: false,

        Login: {
          Name: '',
          Password: ''
        },

        rules: { Login: {} },

        rules0: {
          Login: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入用户名')
            ],

            Password: [
              v => (v && v.length > 0 ? true : '请输入密码')
            ]
          }
        }
      }
    },

    computed: {
      ...mapGetters([
          'alertArea',
          'alertType',
          'alertMsg',
          'token'
      ])
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Me().then(data => {
            context.setToken(data);
          }, err => {
            context.setToken({});
          })
      },

      login() {
        this.rules.Login = this.rules0.Login;

        // 前面对rules的赋值必须在next tick才能生效，validateForm才能正常工作
        this.$nextTick(_ => {
          if (!this.validateForm('Login_')) {
            return;
          }

          api.Login({
            Name: this.Login.Name,
            Pass: this.Login.Password
          }).then(data => {
            window.location.reload();
          }, err => {
            ui.alert('用户名或密码不正确');
          });
        });
      },

      logout() {
        api.Logout().then(data => {
          window.location.reload();
        }, err => {
        });
      }
    }
  }
</script>

<style lang="stylus">
  @import './stylus/main'

  .login-ui
    background-color: #2196f3
    width: 100%
    height: 100%

  .login-content
    position: relative
    top: 20%

  .login-box
    position: relative
    z-index: 999

  .login-footer
    width: 100%
    height: 205px
    position: relative
    bottom: 85px
    left: 0
    background: url("/public/loginbottom.png") no-repeat
    background-size: 100% 205px
    z-index: 1
</style>
