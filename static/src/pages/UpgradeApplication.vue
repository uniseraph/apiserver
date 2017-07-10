<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;应用升级
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
        请选择应用模板
        <v-spacer></v-spacer>
        <v-text-field
            append-icon="search"
            label="模板名称"
            hide-details
            v-model="Keyword"
            @keydown.enter.native="getDataFromApi"
          ></v-text-field>
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
            <td><v-radio label="" v-model="ApplicationTemplateId" :value="props.item.Id"></v-radio></td>
            <td><router-link :to="'/templates/' + props.item.Id">{{ props.item.Title }}</router-link></td>
            <td>{{ props.item.Name }}</td>
            <td>{{ props.item.Version }}</td>
            <td>{{ props.item.Description }}</td>
            <td>{{ props.item.UpdatedTime | formatDateTime }}</td>
            <td>{{ props.item.UpdaterName }}</td>
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
            <v-flex xs12 mt-4 class="text-md-center">
              <v-btn class="orange darken-2 white--text" @click.native="save">
                <v-icon light left>save</v-icon>升级应用
              </v-btn>     
            </v-flex>
            <v-flex xs3>
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
          { text: '应用名称', sortable: false, left: true },
          { text: '应用ID', sortable: false, left: true },
          { text: '应用版本', sortable: false, left: true },
          { text: '说明', sortable: false, left: true },
          { text: '更新时间', sortable: false, left: true },
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

        ApplicationTemplateId: null,

        Id: this.$route.params.id,
        Title: '',
        Name: '',
        Version: '',
        Description: '',

        rules: {},

        rules0: {
          ApplicationTemplateId: [
            v => (v && v.length > 0 ? true : '请选择应用模板')
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

    methods: {
      init() {
        api.Application(this.Id).then(data => {
          this.Id = data.Id;
          this.Title = data.Title;
          this.Name = data.Name;
          this.Version = data.Version;
          this.Description = data.Description;
        });

        this.getDataFromApi();
      },

      goback() {
        this.$router.go(-1);
      },

      getDataFromApi() {
        let params = {
          Keyword: this.Keyword,
          Name: this.Name,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        api.Templates(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      save() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.ApplicationTemplateId) {
            ui.alert('请选择应用模板');
            return;
          }

          if (!this.validateForm()) {
            return;
          }

          let params = {
            Id: this.Id,
            ApplicationTemplateId: this.ApplicationTemplateId
          };

          api.UpgradeApplication(params).then(data => {
            ui.alert('升级应用成功', 'success');
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
