<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;{{ PoolName }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ApplicationTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;{{ ServiceTitle }}&nbsp;&nbsp;/&nbsp;&nbsp;容器列表
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="RestartConfirmDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>提示</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>你确认要重启容器{{ SelectedContainer.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="restartContainer">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RestartConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-layout row justify-center>
        <v-dialog v-model="SSHInfoDlg" persistent width="540">
          <v-card>
            <v-card-row>
              <v-card-title>{{ SelectedContainer.Name }}登录信息</v-card-title>
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
      <v-data-table
        :headers="headers"
        :items="items"
        :total-items="totalItems"
        :pagination.sync="pagination"
        hide-actions
        class="containers-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td>{{ props.item.Id }}</td>
          <td>{{ props.item.Name }}</td>
          <td :class="applicationClass(props.item.Status)">{{ applicationStatus(props.item.Status) }}</td>
          <td>{{ props.item.IP }}</td>
          <td>{{ props.item.Node ? props.item.Node.Name : '' }}</td>
          <td>{{ props.item.Node ? props.item.Node.IP : '' }}</td>
          <td>
            <v-btn outline small icon class="green green--text" @click.native="displaySSHInfo(props.item)" title="登录信息">
              <v-icon>lock_outline</v-icon>
            </v-btn>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRestart(props.item)" title="重启容器">
              <v-icon>refresh</v-icon>
            </v-btn>
          </td>
        </template>
      </v-data-table>
      <div class="text-xs-center pt-2 pb-2">
        <v-pagination v-model="pagination.page" :length="Math.ceil(pagination.totalItems / pagination.rowsPerPage)"></v-pagination>
      </div>
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
          { text: '容器ID', sortable: false, left: true },
          { text: '容器名', sortable: false, left: true },
          { text: '状态', sortable: false, left: true },
          { text: 'IP', sortable: false, left: true },
          { text: '宿主机名', sortable: false, left: true },
          { text: '宿主机IP', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],
        totalItems: 0,
        pagination: { 
          rowsPerPage: this.$route.query ? (this.$route.query.PageSize ? parseInt(this.$route.query.PageSize) : 20) : 20, 
          totalItems: 0, 
          page: this.$route.query ? (this.$route.query.Page ? parseInt(this.$route.query.Page) : 1) : 1, 
          sortBy: this.$route.query ? (this.$route.query.SortBy || '') : '', 
          descending: this.$route.query ? (this.$route.query.Desc || false) : false 
        },

        ApplicationId: this.$route.params.applicationId,
        ServiceName: this.$route.params.serviceName,
        PoolName: this.$route.params.poolName,
        ApplicationTitle: this.$route.params.applicationTitle,
        ServiceTitle: this.$route.params.serviceTitle,

        Keyword: this.$route.query ? (this.$route.query.Keyword || '') : '',

        RestartConfirmDlg: false,
        SelectedContainer: {},

        SSHInfoDlg: false,
        SSHInfo: {}
      }
    },

    watch: {
      'pagination.rowsPerPage': 'paginationChanged',
      'pagination.page': 'paginationChanged',
      'pagination.sortBy': 'paginationChanged',
      'pagination.descending': 'paginationChanged'
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        this.getDataFromApi();
      },

      goback() {
        this.$router.go(-1);
      },

      paginationChanged(v, o) {
        if (v != o) {
          this.getDataFromApi();
        }
      },

      displaySSHInfo(container) {
        api.ContainerSSHInfo(container.Id).then(data => {
          this.SSHInfo = data;
          this.SSHInfoDlg = true;
        });
      },

      confirmBeforeRestart(container) {
        this.SelectedContainer = container;
        this.RestartConfirmDlg = true;
      },

      restartContainer() {
        this.RestartConfirmDlg = false;
        api.RestartContainer(this.SelectedContainer.Id).then(data => {
          ui.alert('容器重启成功', 'success');
          this.getDataFromApi();
        });
      },

      getDataFromApi() {
        let params = {
          Id: this.ApplicationId,
          ServiceName: this.ServiceName,
          Keyword: this.Keyword,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        api.Containers(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      selectAll(i) {
        this.$refs[i].$refs.input.select();
      }
    }
  }
</script>

<style lang="stylus">
.containers-table
  tr
    .btn
      visibility: hidden
  tr:hover
    .btn
      visibility: visible

.dialog
  .input-group
    &__details
      min-height: 22px
</style>
