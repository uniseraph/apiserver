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
                <v-text-field ref="UpdateDir_Name" required v-model="SelectedDir.Name" single-line :rules="rules.Dir.Name"></v-text-field>
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
              <v-card-title>新建“{{ SelectedDir.Name }}”的子目录</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field ref="NewDir_Name" required v-model="NewDir.Name" single-line :rules="rules.Dir.Name"></v-text-field>
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
        <v-layout row justify-center style="margin-right:0;">
          <v-dialog v-model="CreateValueDlg">
            <v-btn class="primary white--text" slot="activator"><v-icon light>add</v-icon>新增参数</v-btn>
            <v-card>
              <v-card-row>
                <v-card-title>新增参数</v-card-title>
              </v-card-row>
              <v-card-row>
                <v-card-text>
                  <v-text-field ref="NewValue_Name" label="名称" required v-model="NewValue.Name" :rules="rules.Value.Name"></v-text-field>
                  <v-text-field ref="NewValue_Value" label="默认值" required v-model="NewValue.Value" :rules="rules.Value.Value"></v-text-field>
                  <v-text-field label="描述" v-model="NewValue.Description"></v-text-field>
                </v-card-text>
              </v-card-row>
              <v-card-row actions>
                <v-btn class="blue--text darken-1" flat @click.native="createValue">确认</v-btn>
                <v-btn class="blue--text darken-1" flat @click.native="CreateValueDlg = false">取消</v-btn>
              </v-card-row>
            </v-card>
          </v-dialog>
        </v-layout>
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
          <td><router-link :to="'/env/' + props.item.Id + '/detail'">{{ props.item.Name }}</router-link></td>
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
        treeData: { nodeData: [], currentNodeId: null },
        headers: [
          { text: '参数ID', sortable: false, left: true },
          { text: '参数名', sortable: false , left: true},
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
        NewDir: { Name: '' },

        RemoveValueConfirmDlg: false,
        SelectedValue: {},

        CreateValueDlg: false,
        NewValue: { Name: '', Value: '', Description: '' },

        rules: {
          Dir: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入目录名')
            ]
          },

          Value: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入参数名')
            ],

            Value: [
              v => (v && v.length > 0 ? true : '请输入默认值')
            ]
          },
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
      init(selectedDirId) {
        let state = this.$refs.tree.getState();
        api.EnvDirs().then(data => {
          let nodeData = [{
            id: '0',
            label: '全部',
            open: true,
            visible: true,
            checked: false,
            children: conv2NodeData('id', data)
          }];

          this.treeData = this.$refs.tree.createTreeData(nodeData, state, selectedDirId);

          if (selectedDirId) {
            let node = this.$refs.tree.getNodeById(selectedDirId);
            this.nodeClicked(node);
          }
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

      validateForm(refPrefix) {
        for (let f in this.$refs) {
          if (f.indexOf(refPrefix) == 0) {
            let e = this.$refs[f];
            if (e.errorBucket && e.errorBucket.length > 0) {
              return false;
            }
          }
        }

        return true;
      },

      createDir() {
        if (!this.validateForm('NewDir_')) {
          return;
        }

        this.CreateDirDlg = false;
        let params = {
          Name: this.NewDir.Name,
          ParentId: this.SelectedDir.Id != '0' ? this.SelectedDir.Id : null
        };
        api.CreateEnvDir(params).then(data => {
          this.init(data.Id);
        });
      },

      updateDir() {
        if (!this.validateForm('UpdateDir_')) {
          return;
        }

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
          this.init(this.SelectedDir.ParentId);
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

      createValue() {
        if (!this.validateForm('NewValue_')) {
          return;
        }

        this.CreateValueDlg = false;
        let params = {
          Name: this.NewValue.Name,
          Value: this.NewValue.Value,
          Description: this.NewValue.Description,
          DirId: this.SelectedDir.Id != '0' ? this.SelectedDir.Id : null
        };
        api.CreateEnvValue(params).then(data => {
          this.getDataFromApi();
        });
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

  function conv2NodeData(pid, list) {
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
        a.children = conv2NodeData(a.id, e.Children);
      }
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
