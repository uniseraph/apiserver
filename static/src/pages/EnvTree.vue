<template>
  <div>
    <v-card-title style="padding-left:0;">
      <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
      &nbsp;&nbsp;参数目录&nbsp;&nbsp;/&nbsp;&nbsp;{{ TreeName }}
      <v-spacer></v-spacer>
    </v-card-title>
    <div>
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
                  <v-alert 
                    v-if="alertArea==='UpdateDirDlg'"
                    v-bind:success="alertType==='success'" 
                    v-bind:info="alertType==='info'" 
                    v-bind:warning="alertType==='warning'" 
                    v-bind:error="alertType==='error'" 
                    v-model="alertMsg" 
                    dismissible>{{ alertMsg }}</v-alert>
                  <v-card-row>
                    <v-card-title>修改目录名</v-card-title>
                  </v-card-row>
                  <v-card-row>
                    <v-card-text>
                      <v-text-field 
                        v-model="SelectedDir.Name" 
                        ref="UpdateDir_Name" 
                        required 
                        :rules="rules.Dir.Name"
                        @input="rules.Dir.Name = rules0.Dir.Name"
                      ></v-text-field>
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
                  <v-alert 
                    v-if="alertArea==='CreateDirDlg'"
                    v-bind:success="alertType==='success'" 
                    v-bind:info="alertType==='info'" 
                    v-bind:warning="alertType==='warning'" 
                    v-bind:error="alertType==='error'" 
                    v-model="alertMsg" 
                    dismissible>{{ alertMsg }}</v-alert>
                  <v-card-row>
                    <v-card-title>新建“{{ SelectedDir.Name }}”的子目录</v-card-title>
                  </v-card-row>
                  <v-card-row>
                    <v-card-text>
                      <v-text-field 
                        v-model="NewDir.Name"
                        ref="NewDir_Name" 
                        required
                        :rules="rules.Dir.Name"
                        @input="rules.Dir.Name = rules0.Dir.Name"
                      ></v-text-field>
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
              <tree ref="tree" :options="treeOptions" :treeData="treeData" @node-click="nodeClicked"></tree>
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
                hide-details
                v-model="Keyword"
                @keydown.enter.native="getDataFromApi"
              ></v-text-field>
              <v-spacer></v-spacer>
              <v-layout row justify-center style="margin-right:0;">
                <v-dialog v-model="CreateValueDlg">
                  <v-btn class="primary white--text" slot="activator"><v-icon light>add</v-icon>新增参数</v-btn>
                  <v-card>
                    <v-alert 
                      v-if="alertArea==='CreateValueDlg'"
                      v-bind:success="alertType==='success'" 
                      v-bind:info="alertType==='info'" 
                      v-bind:warning="alertType==='warning'" 
                      v-bind:error="alertType==='error'" 
                      v-model="alertMsg" 
                      dismissible>{{ alertMsg }}</v-alert>
                    <v-card-row>
                      <v-card-title>新增参数</v-card-title>
                    </v-card-row>
                    <v-card-row>
                      <v-card-text>
                        <v-text-field 
                          v-model="NewValue.Name" 
                          ref="NewValue_Name" 
                          label="名称" 
                          required 
                          :rules="rules.Value.Name"
                          @input="rules.Value.Name = rules0.Value.Name"
                        ></v-text-field>
                        <v-checkbox label="敏感数据" v-model="NewValue.Mask" dark></v-checkbox>
                        <v-text-field 
                          v-model="NewValue.Value" 
                          ref="NewValue_Value" 
                          label="默认值" 
                          required 
                          :rules="rules.Value.Value" 
                          @input="rules.Value.Value = rules0.Value.Value"
                          rel="NewValue_Value"
                          class="completer-field mt-4"
                        ></v-text-field>
                        <v-text-field 
                          label="说明" 
                          v-model="NewValue.Description" 
                          class="mt-4"
                        ></v-text-field>
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
              hide-actions
              class="values-table elevation-1"
              no-data-text=""
            >
              <template slot="items" scope="props">
                <td><router-link :to="'/env/trees/values/' + props.item.Id + '/' + TreeId">{{ props.item.Name }}</router-link></td>
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
    </div>
  </div>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from '../api/api'
  import jQuery from 'jquery'
  import caret from '../caret'
  import completer from '../completer'
  import * as ui from '../util/ui'
  import Tree from '../components/tree/tree.vue'

  export default {
    data() {
      return {
        treeOptions: {},
        treeData: { nodeData: [], currentNodeId: null },
        headers: [
          { text: '参数名', sortable: false , left: true},
          { text: '默认值', sortable: false, left: true },
          { text: '说明', sortable: false, left: true },
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

        TreeId: this.$route.params.id,
        TreeName: this.$route.params.name,

        SelectedDir: {},
        Keyword: '',

        UpdateDirDlg: false,

        RemoveDirConfirmDlg: false,
        RemoveDirDisabled: true,

        CreateDirDlg: false,
        NewDir: { Name: '' },

        RemoveValueConfirmDlg: false,
        SelectedValue: {},

        CreateValueDlg: false,
        NewValue: { Name: '', Mask: false, Value: '', Description: '' },

        rules: { Dir: {}, Value: {} },

        rules0: {
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
        'pagination.rowsPerPage': 'paginationChanged',
        'pagination.page': 'paginationChanged',
        'pagination.sortBy': 'paginationChanged',
        'pagination.descending': 'paginationChanged',

        UpdateDirDlg(v) {
          (v ? ui.showAlertAt('UpdateDirDlg') : ui.showAlertAt())
        },

        CreateDirDlg(v) {
          (v ? ui.showAlertAt('CreateDirDlg') : ui.showAlertAt())
        },

        CreateValueDlg(v) {
          (v ? ui.showAlertAt('CreateValueDlg') : ui.showAlertAt())

          if (v) {
            this.initCompleters();
          }
        }
    },

    computed: {
      ...mapGetters([
          'alertArea',
          'alertType',
          'alertMsg'
      ])
    },

    mounted() {
      this.init(this.$route.query ? this.$route.query.DirId : null);
    },

    methods: {
      init(selectedDirId) {
        let state = this.$refs.tree.getState();
        api.EnvDirs({ TreeId: this.TreeId }).then(data => {
          let nodeData = conv2NodeData(null, [ data ]);

          if (nodeData.length > 0) {
            nodeData[0].open = true;

            if (!selectedDirId) {
              selectedDirId = nodeData[0].id;
            }
          }

          this.treeData = this.$refs.tree.createTreeData(nodeData, state, selectedDirId);

          this.$nextTick(_ => {
            let node = this.$refs.tree.getNodeById(selectedDirId);
            if (node) {
              this.nodeClicked(node);
            }
          });
        });
      },

      initCompleters() {
        this.$nextTick(function() {
            let that = this;
            jQuery('.completer-field').find('input').completer({
              url: this.$axios.defaults.baseURL + '/envs/values/search?TreeId=' + this.TreeId,
              completeSuggestion: function(e, v) {
                let rel = e.parents('.completer-field').attr('rel');
                Object.keys(that.$refs).forEach(k => {
                  if (k != rel) {
                    return;
                  }

                  let r = that.$refs[k];
                  if (Array.isArray(r)) {
                    r = r[0];
                  }

                  r.value = v;
                  r.inputValue = v;
                });
              }
            });
          });
      },

      goback() {
        this.$router.go(-1);
      },

      paginationChanged(v, o) {
        if (v != o) {
          this.getDataFromApi();
        }
      },

      nodeClicked(node) {
        this.Keyword = '';

        if (this.SelectedDir.Id == node.id && this.SelectedDir.ParentId) {
          this.UpdateDirDlg = true;
        } else {
          this.SelectedDir = { Id: node.id, Name: node.label, ParentId: node.parentId };
          this.getDataFromApi();
        }

        if (!this.SelectedDir.ParentId) {
          this.RemoveDirDisabled = true;
        } else {
          this.RemoveDirDisabled = false;
        }
      },

      createDir() {
        this.rules.Dir = this.rules0.Dir;
        this.$nextTick(_ => {
          if (!this.validateForm('NewDir_')) {
            return;
          }

          let params = {
            Name: this.NewDir.Name,
            ParentId: this.SelectedDir.Id,
            TreeId: this.TreeId
          };

          api.CreateEnvDir(params).then(data => {
            this.CreateDirDlg = false;
            this.init(data.Id);
          });
        });
      },

      updateDir() {
        this.rules.Dir = this.rules0.Dir;
        this.$nextTick(_ => {
          if (!this.validateForm('UpdateDir_')) {
            return;
          }

          api.UpdateEnvDir(this.SelectedDir).then(data => {
            this.UpdateDirDlg = false;
            this.init(data.Id);
          });
        });
      },

      confirmBeforeRemoveDir() {
        this.RemoveDirConfirmDlg = true;
      },

      removeDir() {
        this.RemoveDirConfirmDlg = false;
        api.RemoveEnvDir(this.SelectedDir.Id).then(data => {
          this.init(this.SelectedDir.ParentId);
        })
      },

      getDataFromApi() {
        let params = {
          TreeId: this.TreeId,
          DirId: this.SelectedDir.ParentId ? this.SelectedDir.Id : '', 
          Name: this.Keyword,
          PageSize: this.pagination.rowsPerPage, 
          Page: this.pagination.page
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        api.EnvValues(params).then(data => {
          this.pagination.totalItems = data.Total;
          this.pagination.page = data.Page;
          this.items = data.Data;
          this.totalItems = data.Total;
        })
      },

      createValue() {
        this.rules.Value = this.rules0.Value;
        this.$nextTick(_ => {
          if (!this.validateForm('NewValue_')) {
            return;
          }

          let params = {
            Name: this.NewValue.Name,
            Mask: this.NewValue.Mask,
            Value: this.NewValue.Value,
            Description: this.NewValue.Description,
            DirId: this.SelectedDir.Id,
            TreeId: this.TreeId
          };

          api.CreateEnvValue(params).then(data => {
            this.CreateValueDlg = false;
            this.getDataFromApi();
          });
        });
      },

      confirmBeforeRemoveValue(v) {
        this.SelectedValue = v;
        this.RemoveValueConfirmDlg = true;
      },

      removeValue() {
        this.RemoveValueConfirmDlg = false;
        api.RemoveEnvValue(this.SelectedValue.Id).then(data => {
          this.getDataFromApi();
        })
      }
    },

    components: {
      Tree
    }
  }

  function conv2NodeData(pnode, list) {
    let arr = [];
    for (let e of list) {
      let a = {
        id: e.Id,
        label: e.Name,
        open: false,
        visible: true,
        checked: false,
        parentId: pnode ? pnode.id : null,
        parentNode: pnode
      };

      arr.push(a);

      if (e.Children) {
        a.children = conv2NodeData(a, e.Children);
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
