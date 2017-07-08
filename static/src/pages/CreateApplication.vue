<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;新增应用
      <v-spacer></v-spacer>
    </v-card-title>
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
            <td><v-radio label="" v-model="ApplicationTemplateId" :value="props.item.Id" @change="selectTemplate(props.item)"></v-radio></td>
            <td><router-link :to="'/templates/' + props.item.Id">{{ props.item.Title }}</router-link></td>
            <td>{{ props.item.Name }}</td>
            <td>{{ props.item.Version }}</td>
            <td>{{ props.item.Description }}</td>
            <td>{{ props.item.UpdatedTime | formatDate }}</td>
            <td>{{ props.item.Updater.Name }}</td>
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
            <v-flex xs2>
              <v-subheader>目标集群<span class="required-star">*</span></v-subheader>
            </v-flex>
            <v-flex xs3>
              <v-select
                :items="PoolList"
                item-text="Name"
                item-value="Id"
                v-model="PoolId"
                dark
              ></v-select>
            </v-flex>
            <v-flex xs2>
            </v-flex>
            <v-flex xs2>
              <v-subheader>应用名称<span class="required-star">*</span></v-subheader>
            </v-flex>
            <v-flex xs3>
              <v-text-field
                ref="Title"
                v-model="Title"
                required
                :rules="rules.Title"
                @input="rules.Title = rules0.Title"
              ></v-text-field>
            </v-flex>
            <v-flex xs2>
              <v-subheader>说明</v-subheader>
            </v-flex>
            <v-flex xs10>
              <v-text-field
                v-model="Description"
              ></v-text-field>
            </v-flex>
            <v-flex xs12 mt-4 class="text-md-center">
              <v-btn class="orange darken-2 white--text" @click.native="save">
                <v-icon light left>save</v-icon>发布应用
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
          sortBy: this.$route.query ? (this.$route.query.SortBy ? parseInt(this.$route.query.SortBy) : null) : null, 
          descending: this.$route.query ? (this.$route.query.Desc ? parseInt(this.$route.query.Desc) : false) : false 
        },

        PoolList: [],
        Keyword: '',

        ApplicationTemplateId: null,
        PoolId: this.$route.params.poolId,
        Title: '',
        Description: '',

        rules: {},

        rules0: {
          ApplicationTemplateId: [
            v => (v && v.length > 0 ? true : '请选择应用模板')
          ],
          PoolId: [
            v => (v && v.length > 0 ? true : '请选择集群')
          ],
          Name: [
            v => (v && v.length > 0 ? true : '请填写应用名称')
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
        this.getDataFromApi();
        
        api.Pools().then(data => {
          this.PoolList = data;
          this.PoolId = this.$route.params.poolId;
        });
      },

      goback() {
        this.$router.go(-1);
      },

      getDataFromApi() {
        let params = {
          Keyword: this.Keyword,
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

      selectTemplate(template) {
        this.Title = template.Title;
        this.Description = template.Description;
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
            ApplicationTemplateId: this.ApplicationTemplateId,
            PoolId: this.PoolId,
            Title: this.Title,
            Description: this.Description
          };

          api.Applications(params).then(data => {
            ui.alert('发布应用成功', 'success');
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
