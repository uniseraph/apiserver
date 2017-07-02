<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;用户列表&nbsp;&nbsp;/&nbsp;&nbsp;重置密码&nbsp;&nbsp;/&nbsp;&nbsp;{{ Name }}
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs2>
            <v-subheader>新密码</v-subheader>
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
            <v-subheader>再输一次</v-subheader>
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
          <v-flex xs2>
          </v-flex>
          <v-flex xs10 mt-4 class="text-md-left">
            <v-btn class="orange darken-2 white--text" @click.native="save">
              <v-icon light left>save</v-icon>确认修改
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
        Password1: '',
        Password2: '',

        rules: {},

        rules0: {
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
        api.User(this.$route.params.id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
        })
      },

      goback() {
        router.go(-1);
      },

      save() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.validateForm()) {
            return;
          }

          api.ResetPassword({
            Id: this.Id,
            Pass: this.Password1
          }).then(data => {
            ui.alert('密码修改成功', 'success');
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
