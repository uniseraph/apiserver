<template>
  <v-card>
    <v-card-title>
      应用管理
      <v-spacer></v-spacer>
      <v-select
          :items="PoolList"
          item-text="Name"
          item-value="Id"
          v-model="PoolId"
          dark
          @input="poolChanged"
        ></v-select>
      <v-text-field
          append-icon="search"
          label="应用名称"
          hide-details
          v-model="Keyword"
          @keydown.enter.native="getDataFromApi"
          class="ml-4"
        ></v-text-field>
      <router-link :to="'/applications/create/' + PoolId">
        <v-btn class="primary white--text ml-4"><v-icon light>add</v-icon>新增应用</v-btn>
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
              <v-card-text>你确认要删除应用{{ SelectedApplication.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeApplication">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveConfirmDlg = false">取消</v-btn>
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
        class="applications-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td><router-link :to="'/applications/' + props.item.Id + '/' + encodeURIComponent(PoolMap[props.item.PoolId])">{{ props.item.Title }}</router-link></td>
          <td>{{ props.item.Name }}</td>
          <td>{{ props.item.Version }}</td>
          <td :class="applicationClass(props.item.Status)">{{ applicationStatus(props.item.Status) }}</td>
          <td>{{ props.item.UpdatedTime | formatDateTime }}</td>
          <td>{{ props.item.UpdaterName }}</td>
          <td>
            <v-btn v-if="props.item.Status==='running'" outline small icon class="red red--text" @click.native="stopApplication(props.item)" title="停止应用">
              <v-icon>pause</v-icon>
            </v-btn>
            <v-btn v-if="props.item.Status==='stopped'" outline small icon class="blue blue--text" @click.native="startApplication(props.item)" title="启动应用">
              <v-icon>play_arrow</v-icon>
            </v-btn>
            <v-btn v-if="props.item.Status==='stopped'" outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除应用">
              <v-icon>close</v-icon>
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
          { text: '应用名称', sortable: false, left: true },
          { text: '应用ID', sortable: false, left: true },
          { text: '应用版本', sortable: false, left: true },
          { text: '状态', sortable: false, left: true },
          { text: '更新时间', sortable: false, left: true },
          { text: '操作人', sortable: false, left: true },
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

        PoolList: [],
        PoolMap: {},
        PoolId: this.$route.query ? (this.$route.query.PoolId || '') : '', 
        Keyword: this.$route.query ? (this.$route.query.Keyword || '') : '',

        RemoveConfirmDlg: false,
        SelectedApplication: {}
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
        api.Pools().then(data => {
          this.PoolList = data;
          if (!this.PoolId && data.length > 0) {
            this.PoolId = data[0].Id;
          }

          this.PoolMap = {};
          for (let p of data) {
            this.PoolMap[p.Id] = p.Name;
          }

          this.getDataFromApi();
        })  
      },

      paginationChanged(v, o) {
        if (v != o) {
          this.getDataFromApi();
        }
      },

      poolChanged(id) {
        this.getDataFromApi();
      },

      getDataFromApi() {
        if (!this.PoolId || this.PoolId.length == 0) {
          return;
        }

        let params = {
          PoolId: this.PoolId,
          Keyword: this.Keyword,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        api.Applications(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      startApplication(application) {
        api.StartApplication(application.Id).then(data => {
          this.getDataFromApi();
        })
      },

      stopApplication(application) {
        api.StopApplication(application.Id).then(data => {
          this.getDataFromApi();
        })
      },

      confirmBeforeRemove(application) {
        this.SelectedApplication = application;
        this.RemoveConfirmDlg = true;
      },

      removeApplication() {
        this.RemoveConfirmDlg = false;
        api.RemoveApplication(this.SelectedApplication.Id).then(data => {
          this.getDataFromApi();
        })
      }
    }
  }
</script>

<style lang="stylus">
.applications-table
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
