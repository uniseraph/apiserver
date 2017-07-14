<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;应用回滚
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs2>
            <v-subheader>应用名称</v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Title"
              readonly
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs2>
            <v-subheader>应用ID</v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Name"
              readonly
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
            <v-subheader>应用版本</v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Version"
              readonly
            ></v-text-field>
          </v-flex>
          <v-flex xs2>
          </v-flex>
          <v-flex xs2>
            <v-subheader>说明</v-subheader>
          </v-flex>
          <v-flex xs3>
            <v-text-field
              v-model="Description"
              readonly
            ></v-text-field>
          </v-flex>
        </v-layout>
      </v-container>
    </div>
    <v-divider></v-divider>
    <div class="ml-4 mr-4">
      <v-card-title>
        发布记录
        <v-spacer></v-spacer>
      </v-card-title>
      <div>
        <v-data-table
          :headers="headers"
          :items="items"
          :total-items="totalItems"
          :pagination.sync="pagination"
          hide-actions
          class="templates-table elevation-1"
          no-data-text=""
        >
          <template slot="items" scope="props">
            <td><v-radio label="" v-model="DeploymentHistoryId" :value="props.item.Id"></v-radio></td>
            <td>{{ props.item.Version }}</td>
            <td v-if="props.item.OperationType == 'create'">新增</td>
            <td v-if="props.item.OperationType == 'upgrade'">升级</td>
            <td v-if="props.item.OperationType == 'rollback'">回滚</td>
            <td>{{ props.item.CreatedTime | formatDateTime }}</td>
            <td>{{ props.item.CreatorName }}</td>
          </template>
        </v-data-table>
        <div class="text-xs-center pt-2 pb-2">
          <v-pagination v-model="pagination.page" :length="Math.ceil(pagination.totalItems / pagination.rowsPerPage)"></v-pagination>
        </div>
      </div>
      <v-divider></v-divider>
      <div>
        <v-container fluid>
          <v-layout row wrap>
            <v-flex xs12>
              <v-alert 
                    v-if="alertArea==='RollbackApplication'"
                    v-bind:success="alertType==='success'" 
                    v-bind:info="alertType==='info'" 
                    v-bind:warning="alertType==='warning'" 
                    v-bind:error="alertType==='error'" 
                    v-model="alertMsg" 
                    dismissible>{{ alertMsg }}</v-alert>
            </v-flex>
            <v-flex v-if="!Submitting" xs12 mt-4 class="text-md-center">
              <v-btn class="orange darken-2 white--text" @click.native="save">
                <v-icon light left>save</v-icon>回滚应用
              </v-btn>     
            </v-flex>
            <v-flex v-if="Submitting" xs12 mt-4 class="text-md-center">
              <v-progress-linear v-bind:indeterminate="true"></v-progress-linear>
            </v-flex>
          </v-layout>
        </v-container>
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
          { text: '选择', sortable: false, left: true },
          { text: '应用版本', sortable: false, left: true },
          { text: '发布类型', sortable: false, left: true },
          { text: '更新时间', sortable: false, left: true },
          { text: '操作人', sortable: false, left: true }
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
        Keyword: '',

        DeploymentHistoryId: null,

        Id: this.$route.params.id,
        Title: '',
        Name: '',
        Version: '',
        Description: '',

        Submitting: false,

        rules: {},

        rules0: {
          DeploymentHistoryId: [
            v => (v && v.length > 0 ? true : '请选择发布记录')
          ]
        }
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

    destroyed() {
      ui.showAlertAt();
    },

    methods: {
      init() {
        api.Application(this.Id).then(data => {
          this.Id = data.Id;
          this.Title = data.Title;
          this.Name = data.Name;
          this.Version = data.Version;
          this.Description = data.Description;

          this.getDataFromApi();
        });
      },

      goback() {
        this.$router.go(-1);
      },

      getDataFromApi() {
        let params = {
          Id: this.Id,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        api.DeploymentHistory(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      save() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.DeploymentHistoryId) {
            ui.alert('请选择发布记录');
            return;
          }

          if (!this.validateForm()) {
            return;
          }

          let params = {
            Id: this.Id,
            DeploymentHistoryId: this.DeploymentHistoryId
          };

          ui.showAlertAt('RollbackApplication');
          this.Submitting = true;

          api.RollbackApplication(params).then(data => {
            ui.alert('回滚应用成功', 'success');
            this.Submitting = false;
            let that = this;
            setTimeout(() => {
              that.goback();
            }, 1500);
          }, err => {
            this.Submitting = false;
          });
        });
      }
    }
  }

</script>

<style lang="stylus">

</style>
