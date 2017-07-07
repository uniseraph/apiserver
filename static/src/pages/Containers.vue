<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      容器列表 / {{ ServiceTitle }}
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
          <td :class="{ 'green--text': props.item.Status==='running', 'orange--text': props.item.Status==='stopped', 'red--text': props.item.Status!=='running' && props.item.Status!=='stopped' }">{{ props.item.Status==='running' ? '运行中' : (props.item.Status==='stopped' ? '已停止' : '未知') }}</td>
          <td>{{ props.item.Network ? props.item.Network.IP : '' }}</td>
          <td>{{ props.item.StartedTime | formatDateTime }}</td>
          <td>{{ props.item.Node ? props.item.Node.Name : '' }}</td>
          <td>{{ props.item.Node ? props.item.Node.IP : '' }}</td>
          <td>
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
          { text: '启动时间', sortable: false, left: true },
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
          sortBy: this.$route.query ? (this.$route.query.SortBy ? parseInt(this.$route.query.SortBy) : null) : null, 
          descending: this.$route.query ? (this.$route.query.Desc ? parseInt(this.$route.query.Desc) : false) : false 
        },

        ApplicationId: this.$route.params.applicationId,
        ServiceName: this.$route.params.serviceName,
        ServiceTitle: this.$route.params.serviceTitle,

        Keyword: this.$route.query ? (this.$route.query.Keyword ? parseInt(this.$route.query.Keyword) : '') : '',

        RestartConfirmDlg: false,
        SelectedContainer: {}
      }
    },

    watch: {
        pagination: {
          handler(v, o) {
            this.getDataFromApi();
          },

          deep: true
        }
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

      confirmBeforeRestart(container) {
        this.SelectedContainer = container;
        this.RestartConfirmDlg = true;
      },

      restartContainer() {
        this.RestartConfirmDlg = false;
        api.RestartContainer(this.SelectedContainer.Id).then(data => {
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