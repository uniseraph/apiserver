<template>
  <v-layout row>
  <v-flex xs4>
    <v-card>
      <v-card-title>
        参数目录
        <v-spacer></v-spacer>
        <div>
          <v-btn icon class="blue--text text--lighten-2" @click.native="CreateDirDlg = true">
            <v-icon light>add</v-icon>
          </v-btn>
          <v-btn ref="RemoveDir" :disabled="RemoveDirDisabled" icon class="red--text text--lighten-2" @click.native="confirmBeforeRemoveDir">
            <v-icon light>remove</v-icon>
          </v-btn>
        </div>
      </v-card-title>
      <v-layout row justify-center>
        <v-dialog v-model="UpdateDirDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>修改目录名</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field ref="UpdateDirName" v-model="SelectedDir.Name" :rules="rules.Dir.Name"></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="updateDir">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="UpdateDirDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-layout row justify-center>
        <v-dialog v-model="CreateDirDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>新建子目录</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field ref="NewDirName" v-model="NewDirName" :rules="rules.Dir.Name"></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="createDir">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="CreateDirDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-layout row justify-center>
        <v-dialog v-model="RemoveDirConfirmDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>提示</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>你确认要删除目录“{{ SelectedDir.Name }}”吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeDir">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveDirConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <div class="pl-4 pr-4 pb-2">
        <tree ref="tree" :options="treeOptions" :treeData="treeData" @node-click="nodeClicked" />
      </div>
    </v-card>
  </v-flex>
  <v-flex xs8>
    <v-layout row justify-center>
      <v-dialog v-model="RemoveValueConfirmDlg" persistent>
        <v-card>
          <v-card-row>
            <v-card-title>提示</v-card-title>
          </v-card-row>
          <v-card-row>
            <v-card-text>你确认要删除参数{{ SelectedValue.Name }}吗？</v-card-text>
          </v-card-row>
          <v-card-row actions>
            <v-btn class="green--text darken-1" flat="flat" @click.native="removeValue">确认</v-btn>
            <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveValueConfirmDlg = false">取消</v-btn>
          </v-card-row>
        </v-card>
      </v-dialog>
    </v-layout>
    <v-card>
      <v-card-title>
        <v-text-field
          append-icon="search"
          label="参数名称"
          single-line
          hide-details
          v-model="Keyword"
          @keydown.enter.native="getDataFromApi"
        ></v-text-field>
        <v-spacer></v-spacer>

      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :total-items="totalItems"
        :pagination.sync="pagination"
        :search="Keyword"
        hide-actions
        class="values-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td>{{ props.item.Id }}</td>
          <td><router-link :to="'/envs/values/' + props.item.Id + '/detail'">{{ props.item.Name }}</router-link></td>
          <td>{{ props.item.Value }}</td>
          <td>{{ props.item.Description }}</td>
          <td>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemoveValue(props.item)" title="删除参数">
              <v-icon>close</v-icon>
            </v-btn>
          </td>
        </template>
      </v-data-table>
      <div class="text-xs-center pt-2 pb-2">
        <v-pagination v-model="pagination.page" :length="Math.ceil(pagination.totalItems / pagination.rowsPerPage)"></v-pagination>
      </div>
    </v-card>
  </v-flex>
  </v-layout>
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'
  import Tree from '../components/tree/tree.vue'

  export default {
    data() {
      return {
        treeOptions: {},
        treeData: [],
        headers: [
          { text: '参数ID', sortable: false, left: true },
          { text: '参数名称', sortable: false , left: true},
          { text: '默认值', sortable: false, left: true },
          { text: '描述', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],
        totalItems: 0,
        pagination: { rowsPerPage: 2, totalItems: 0, page: 1, sortBy: null, descending: false },

        SelectedDir: { Id: '0', Name: '全部' },
        Keyword: '',

        UpdateDirDlg: false,
        RemoveDirConfirmDlg: false,
        RemoveDirDisabled: true,

        CreateDirDlg: false,
        NewDirName: '',

        RemoveValueConfirmDlg: false,
        SelectedValue: {},

        ParentIdMap: {},

        rules: {
          Dir: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入目录名')
            ]
          }
        }
      }
    },

    watch: {
        pagination: {
          handler() {
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
        this.ParentIdMap = {};
        api.EnvDirs().then(data => {
          let treeData = [{
            id: '0',
            label: '全部',
            open: true,
            visible: true,
            checked: false,
            children: conv2TreeData('id', data, this.ParentIdMap)
          }];

          this.treeData = treeData;

          
        })
      },

      nodeClicked(node) {
        this.Keyword = '';

        if (this.SelectedDir.Id == node.id && node.id != '0') {
          this.UpdateDirDlg = true;
        } else {
          this.SelectedDir = { Id: node.id, Name: node.label, ParentId: node.parentId };
          this.getDataFromApi();
        }

        if (this.SelectedDir.Id == '0') {
          this.RemoveDirDisabled = true;
        } else {
          this.RemoveDirDisabled = false;
        }
      },

      createDir() {
        this.CreateDirDlg = false;
        let params = {
          Name: this.NewDirName,
          ParentId: this.SelectedDir.Id != '0' ? this.SelectedDir.Id : null
        };
        api.CreateEnvDir(this.SelectedDir).then(data => {
          this.init();
        });

        this.NewDirName = '';
      },

      updateDir() {
        this.UpdateDirDlg = false;
        api.UpdateEnvDir(this.SelectedDir).then(data => {
          this.init();
        });
      },

      confirmBeforeRemoveDir() {
        this.RemoveDirConfirmDlg = true;
      },

      removeDir() {
        this.RemoveDirConfirmDlg = false;
        api.RemoveEnvDir({ Id: this.SelectedDir.Id }).then(data => {
          this.init();
        })
      },

      getDataFromApi() {
        let params = {
          DirId: this.SelectedDir.Id != '0' ? this.SelectedDir.Id : null, 
          Name: this.Name,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        api.EnvValues(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        })
      },

      confirmBeforeRemoveValue(v) {
        this.SelectedValue = v;
        this.RemoveValueConfirmDlg = true;
      },

      removeValue() {
        this.RemoveValueConfirmDlg = false;
        api.RemoveEnvValue({ Id: this.SelectedValue.Id }).then(data => {
          this.getDataFromApi();
        })
      }
    },

    components: {
      Tree
    }
  }

  function conv2TreeData(pid, list, pidMap) {
    let arr = [];
    for (let e of list) {
      let a = {
        id: e.Id,
        label: e.Name,
        parentId: e.ParentId ? e.ParentId : '0',
        open: false,
        visible: true,
        checked: false
      };
      arr.push(a);
      if (e.Children) {
        a.children = conv2TreeData(a.id, e.Children, pidMap);
      }

      pidMap[a.id] = pid;
    }

    return arr;
  }
</script>

<style lang="stylus">
.values-table
  tr
    .btn
      visibility: hidden
  tr:hover
    .btn
      visibility: visible
</style>
