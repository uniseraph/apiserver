<template>
  <v-card>
    <v-card-title>
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;应用管理&nbsp;&nbsp;/&nbsp;&nbsp;应用升级
      <v-spacer></v-spacer>
    </v-card-title>
    <div class="ml-4 mr-4">
      <v-card-title>
        请选择应用模板
        <v-spacer></v-spacer>
        <v-text-field
            append-icon="search"
            label="模板名称"
            single-line
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
        pagination: { rowsPerPage: 2, totalItems: 0, page: 1, sortBy: null, descending: false },

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
