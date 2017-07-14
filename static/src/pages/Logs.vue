<template>
  <v-card>
    <v-card-title>
      系统日志
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="DetailDlg" persistent width="640">
          <v-card>
            <v-card-row>
              <v-card-text>
                <v-text-field 
                  v-model="Detail"
                  readonly
                  multi-line
                  rows="24"
                  full-width
                  class="log-detail"
                ></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="DetailDlg = false">关闭</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs2>
            <v-select
                :items="ModuleList"
                item-text="Name"
                item-value="Id"
                v-model="Module"
                label="模块"
                dark
                @input="moduleChanged"
              ></v-select>
          </v-flex>
          <v-flex xs2 v-if="Module=='Pool'||Module=='Application'">
            <v-select
                :items="PoolList"
                item-text="Name"
                item-value="Id"
                v-model="PoolId"
                label="集群"
                dark
                @input="poolChanged"
            ></v-select>
          </v-flex>
          <v-flex xs2 v-else>
          </v-flex>
          <v-flex xs2 v-if="Module=='Application'">
            <v-select
                :items="ApplicationList"
                item-text="Title"
                item-value="Id"
                v-model="ApplicationId"
                label="应用"
                dark
            ></v-select>
          </v-flex>
          <v-flex xs2 v-else>
          </v-flex>
          <v-flex xs2>
            <v-select
                :items="OperationList"
                item-text="Name"
                item-value="Id"
                v-model="Operation"
                label="操作"
                dark
              ></v-select>
          </v-flex>
          <v-flex xs2>
            <v-select
                :items="UserList"
                item-text="Name"
                item-value="Id"
                v-model="UserId"
                label="用户"
                dark
              ></v-select>
          </v-flex>
          <v-flex xs2>
            <v-text-field
                placeholder="IP"
                hide-details
                v-model="IP"
              ></v-text-field>
          </v-flex>
           <v-flex xs3 mt-4>
            <v-menu
              lazy
              :close-on-content-click="true"
              v-model="StartTime1Menu"
              transition="v-scale-transition"
              offset-y
              full-width
              :nudge-left="40"
              max-width="290px"
            >
              <v-text-field
                slot="activator"
                label="起始日期"
                v-model="StartTime1"
                prepend-icon="event"
                readonly
              ></v-text-field>
              <v-date-picker v-model="StartTime1" no-title scrollable actions
                :days="constants.DAY_LIST"
                :months="constants.MONTH_LIST"
                >
              </v-date-picker>
            </v-menu>
          </v-flex>
          <v-flex xs2 mt-4>
            <v-select
                :items="constants.TIME_LIST"
                v-model="StartTime2"
                label="起始时间"
                single-line
                dark
            ></v-select>
          </v-flex>
          <v-flex xs3 mt-4>
            <v-menu
              lazy
              :close-on-content-click="true"
              v-model="EndTime1Menu"
              transition="v-scale-transition"
              offset-y
              full-width
              :nudge-left="40"
              max-width="290px"
            >
              <v-text-field
                slot="activator"
                label="截止日期"
                v-model="EndTime1"
                prepend-icon="event"
                readonly
              ></v-text-field>
              <v-date-picker v-model="EndTime1" no-title scrollable actions
                :days="constants.DAY_LIST"
                :months="constants.MONTH_LIST"
                >
              </v-date-picker>
            </v-menu>
          </v-flex>
          <v-flex xs2 mt-4>
            <v-select
                :items="constants.TIME_LIST"
                v-model="EndTime2"
                label="截止时间"
                single-line
                dark
            ></v-select>
          </v-flex>
          <v-flex xs2>
            <v-btn small class="orange darken-2 white--text mt-4" @click.native="getDataFromApi">
              <v-icon light left>search</v-icon>查询
            </v-btn>            
          </v-flex>
        </v-layout>
      </v-container>
      <v-data-table
        :headers="headers"
        :items="items"
        :total-items="totalItems"
        :pagination.sync="pagination"
        hide-actions
        class="logs-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td>{{ props.item.CreatedTime | formatDateTime }}</td>
          <td>{{ props.item.User.Name }}</td>
          <td>{{ props.item.IP }}</td>
          <td>{{ moduleName(props.item.Module) }}</td>
          <td>{{ props.item.Pool.Name }}</td>
          <td>{{ props.item.Application.Title }} ({{ props.item.Application.Name }} {{ props.item.Application.Version }})</td>
          <td>
            {{ operationName(props.item.Module, props.item.Operation) }}
            <v-btn outline small class="green green--text" @click.native="displayDetail(props.item)">
              详细信息
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
          { text: '时间', sortable: false, left: true },
          { text: '用户', sortable: false, left: true },
          { text: 'IP', sortable: false, left: true },
          { text: '模块', sortable: false, left: true },
          { text: '集群', sortable: false, left: true },
          { text: '应用', sortable: false, left: true },
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

        StartTime1Menu: true,
        EndTime1Menu: true,

        PoolList: [],
        ApplicationList: [ { Id: '', Title: '所有应用' } ],
        UserList: [],
        ModuleList: [],
        OperationList: [],

        StartTime1: this.$route.query ? (this.$route.query.StartTime1 || '') : '', 
        StartTime2: this.$route.query ? (this.$route.query.StartTime2 || '00:00') : '00:00', 
        EndTime1: this.$route.query ? (this.$route.query.EndTime1 || '') : '', 
        EndTime2: this.$route.query ? (this.$route.query.EndTime2 || '00:00') : '00:00', 
        PoolId: this.$route.query ? (this.$route.query.PoolId || '') : '', 
        ApplicationId: this.$route.query ? (this.$route.query.ApplicationId || '') : '', 
        UserId: this.$route.query ? (this.$route.query.UserId || '') : '',
        IP: this.$route.query ? (this.$route.query.IP || '') : '',
        Module: this.$route.query ? (this.$route.query.Module || '') : '',
        Operation: this.$route.query ? (this.$route.query.Operation || '') : '',

        DetailDlg: false,
        Detail: ''
      }
    },

    watch: {
        pagination: {
          handler(v, o) {
            if (v.rowsPerPage != o.rowsPerPage || v.page != o.page || v.sortBy != o.sortBy || v.descending != o.descending) {
              this.getDataFromApi();
            }
          },

          deep: true
        }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Pools().then(data => {
          this.PoolList = [{ Id: '', Name: '所有集群' }].concat(data);
        });

        api.Users().then(data => {
          this.UserList = [{ Id: '', Name: '所有用户' }].concat(data);
        });

        this.poolChanged(this.PoolId);

        this.ModuleList = [{ Id: '', Name: '所有模块' }].concat(this.constants.MODULE_LIST);
        this.moduleChanged(this.Module);

        this.getDataFromApi();
      },

      poolChanged(id) {
        let a = [{ Id: '', Title: '所有应用' }];
        if (!id || id.length == 0) {
          this.ApplicationList = a;
          return;
        }

        api.Applications({
          PoolId: id,
          PageSize: 200,
          Page: 1
        }).then(data => {
          this.ApplicationList = a.concat(data.Data);
        });
      },

      moduleChanged(id) {
        let a = [{ Id: '', Name: '所有操作' }];
        if (id && id.length > 0) {
          a = a.concat(this.constants.OPERATION_MAP[id]);
        }

        this.OperationList = a;
      },

      getDataFromApi() {
        let startTime = 0;
        let endTime = 0;

        if (this.StartTime1 && this.StartTime1.length > 0) {
          startTime = Math.floor(this.parseDate(this.StartTime1 + ' ' + this.StartTime2, 'yyyy-MM-dd HH:mm').getTime() / 1000);
        }

        if (this.EndTime1 && this.EndTime1.length > 0) {
          endTime = Math.floor(this.parseDate(this.EndTime1 + ' ' + this.EndTime2, 'yyyy-MM-dd HH:mm').getTime() / 1000);
        }

        let params = {
          StartTime: startTime,
          StartTime1: this.StartTime1,
          StartTime2: this.StartTime2,
          EndTime: endTime,
          EndTime1: this.EndTime1,
          EndTime2: this.EndTime2,
          PoolId: this.PoolId,
          ApplicationId: this.ApplicationId,
          UserId: this.UserId,
          IP: this.IP,
          Module: this.Module,
          Operation: this.Operation,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        api.Logs(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      displayDetail(item) {
        if (item.Detail) {
          this.Detail = JSON.stringify(JSON.parse(item.Detail), null, 4); 
        } else {
          this.Detail = '';
        }
        
        this.DetailDlg = true;
      }
    }
  }
</script>

<style lang="stylus">
.logs-table
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

.input-group--text-field
  &.log-detail
    textarea
      font-size: 12px;
</style>
