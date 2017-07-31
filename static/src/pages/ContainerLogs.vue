<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;{{ PoolName }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ApplicationTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ServiceTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;容器日志
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs2>
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
          <v-flex xs2>
            日志行数：
          </v-flex>
          <v-flex xs2>
            <v-text-field
                required
                v-model="Lines"
                :rules="rules.Lines"
              ></v-text-field>
          </v-flex>
          <v-flex xs3>
            <v-btn small class="orange darken-2 white--text" @click.native="getDataFromApi">
              <v-icon light left>search</v-icon>刷新
            </v-btn>
          </v-flex>
          <v-flex xs12>
            {{ LogText }}
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

        Lines: 200,
        LogText: '',

        rules: {
          Lines: [
            function(o) {
              let v = o ? o.toString() : '';
              return (v && v.length > 0 ? (/^\d+$/.test(v) && parseInt(v) > 0 && parseInt(v) <= 10000 ? true : '日志行数必须为1-10000的整数') : '请输入日志行数')
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
        });

        this.getDataFromApi();
      },

      goback() {
        this.$router.go(-1);
      },

      getDataFromApi() {
        let params = {
          Id: this.ContainerId,
          ShowStdout: true,
          ShowStderr: true,
          Timestamps: true,
          Since: '0',
          Tail: '' + this.Lines
        };

        api.ContainerLogs(params).then(data => {
          this.LogText = data;
        });
      }
    }
  }
</script>

<style lang="stylus">

</style>
