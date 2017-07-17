<template>
  <v-card>
    <v-card-title>
      应用模板
      <v-spacer></v-spacer>
      <v-text-field
          append-icon="search"
          label="模板名称"
          hide-details
          v-model="Keyword"
          @keydown.enter.native="getDataFromApi"
        ></v-text-field>
      <router-link :to="'/templates/create'">
        <v-btn class="primary white--text ml-4"><v-icon light>add</v-icon>新增应用模板</v-btn>
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
              <v-card-text>你确认要删除应用模板{{ SelectedTemplate.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeTemplate">确认</v-btn>
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
        class="templates-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td><router-link :to="'/templates/' + props.item.Id">{{ props.item.Title }}</router-link></td>
          <td>{{ props.item.Name }}</td>
          <td>{{ props.item.Version }}</td>
          <td>{{ props.item.Description }}</td>
          <td>{{ props.item.UpdatedTime | formatDateTime }}</td>
          <td>{{ props.item.UpdaterName }}</td>
          <td>
            <router-link :to="'/templates/copy/' + props.item.Id + '/' + encodeURIComponent('Copy of ' + props.item.Title)">
              <v-btn outline small icon class="green green--text" title="复制应用模板">
                  <v-icon>content_copy</v-icon>
              </v-btn>
            </router-link>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除应用模板">
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
          { text: '说明', sortable: false, left: true },
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

        Keyword: this.$route.query ? (this.$route.query.Keyword || '') : '',

        RemoveConfirmDlg: false,
        SelectedTemplate: {}
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

      paginationChanged(v, o) {
        if (v != o) {
          this.getDataFromApi();
        }
      },
      
      getDataFromApi() {
        let params = {
          Keyword: this.Keyword,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        api.Templates(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        });
      },

      confirmBeforeRemove(template) {
        this.SelectedTemplate = template;
        this.RemoveConfirmDlg = true;
      },

      removeTemplate() {
        this.RemoveConfirmDlg = false;
        api.RemoveTemplate(this.SelectedTemplate.Id).then(data => {
          this.init();
        })
      }
    }
  }
</script>

<style lang="stylus">
.templates-table
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
