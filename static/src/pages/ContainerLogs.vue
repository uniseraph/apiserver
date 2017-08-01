<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;{{ PoolName }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ApplicationTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ServiceTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;容器日志
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="SSHInfoDlg" persistent width="540">
          <v-card>
            <v-card-row>
              <v-card-title>{{ ContainerMap[ContainerId] }}登录信息</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field
                  label="登录命令"
                  ref="SSHInfo_Command"
                  v-model="SSHInfo.Command"
                  readonly
                  @focus="selectAll('SSHInfo_Command')"
                  hint="此登录命令有效期为5分钟"
                  persistent-hint
                ></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="SSHInfoDlg = false">关闭</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs1>
            当前容器：
          </v-flex>
          <v-flex xs3>
            <v-select
                :items="ContainerList"
                item-text="Name"
                item-value="Id"
                v-model="ContainerId"
                dark
                @input="getDataFromApi"
              ></v-select>
          </v-flex>
          <v-flex xs1>
            <v-btn outline small icon class="green green--text" @click.native="displaySSHInfo()" title="登录信息">
              <v-icon>lock_outline</v-icon>
            </v-btn>
          </v-flex>
          <v-flex xs1>
          </v-flex>
          <v-flex xs1>
            日志行数：
          </v-flex>
          <v-flex xs2>
            <v-text-field
                required
                v-model="Lines"
                :rules="rules.Lines"
              ></v-text-field>
          </v-flex>
          <v-flex xs2>
            <v-btn small class="orange darken-2 white--text" @click.native="getDataFromApi">
              <v-icon light left>search</v-icon>刷新
            </v-btn>
          </v-flex>
          <v-flex xs12 mt-2>
            <v-text-field 
                  v-model="LogText"
                  readonly
                  multi-line
                  :rows="Lines"
                  full-width
                  class="log-field"
                ></v-text-field>
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
        ApplicationId: this.$route.params.applicationId,
        ServiceName: this.$route.params.serviceName,
        ContainerId: this.$route.params.containerId,
        PoolName: this.$route.params.poolName,
        ApplicationTitle: this.$route.params.applicationTitle,
        ServiceTitle: this.$route.params.serviceTitle,

        ContainerList: [],
        ContainerMap: {},

        Lines: 200,
        LogText: '',

        SSHInfoDlg: false,
        SSHInfo: {},

        rules: {
          Lines: [
            function(o) {
              let v = o ? o.toString() : '';
              return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 && parseInt(v) <= 9999 ? true : '日志行数必须为1-9999的整数') : '请输入日志行数')
            }
          ]
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        let params = {
          Id: this.ApplicationId,
          ServiceName: this.ServiceName,
          PageSize: 200, 
          Page: 1
        };

        api.Containers(params).then(data => {
          this.ContainerList = data.Data;

          this.ContainerMap = {};
          for (let c of data) {
            this.ContainerMap[c.Id] = c.Name;
          }
        });

        this.getDataFromApi();
      },

      goback() {
        this.$router.go(-1);
      },

      getDataFromApi() {
        let lines = parseInt(this.Lines);
        if (isNaN(lines) || lines <= 0) {
          lines = 200;
        }

        if (lines > 9999) {
          lines = 9999;
        }

        this.Lines = lines;

        let params = {
          Id: this.ContainerId,
          ShowStdout: true,
          ShowStderr: true,
          Timestamps: false,
          Since: '0',
          Tail: '' + this.Lines
        };

        api.ContainerLogs(params).then(data => {
          this.LogText = data;
        });
      },

      displaySSHInfo() {
        api.ContainerSSHInfo(this.ContainerId).then(data => {
          this.SSHInfo = data;
          this.SSHInfoDlg = true;
        });
      },
    }
  }
</script>

<style lang="stylus">
.input-group--text-field.log-field textarea
  font-size: 12px

.dialog
  .input-group
    &__details
      min-height: 22px
</style>
