<template>
  <v-card>
    <v-card-title>
      参数目录
      <v-spacer></v-spacer>
      <v-layout row justify-center style="margin-right:0;">
        <v-dialog v-model="CreateTreeDlg">
          <v-btn class="primary white--text" slot="activator"><v-icon light>add</v-icon>新增目录</v-btn>
          <v-card>
            <v-alert 
              v-if="alertArea==='CreateTreeDlg'"
              v-bind:success="alertType==='success'" 
              v-bind:info="alertType==='info'" 
              v-bind:warning="alertType==='warning'" 
              v-bind:error="alertType==='error'" 
              v-model="alertMsg" 
              dismissible>{{ alertMsg }}</v-alert>
            <v-card-row>
              <v-card-title>新增目录</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field 
                  v-model="NewTree.Name"
                  ref="NewTree_Name" 
                  label="名称" 
                  required
                  :rules="rules.Create.Name"
                  @input="rules.Create.Name = rules0.Create.Name"
                ></v-text-field>
                <v-text-field 
                  v-model="NewTree.Description" 
                  label="说明" 
                  class="mt-4"
                ></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="createTree">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="CreateTreeDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="UpdateTreeDlg" persistent>
          <v-card>
            <v-alert 
              v-if="alertArea==='UpdateTreeDlg'"
              v-bind:success="alertType==='success'" 
              v-bind:info="alertType==='info'" 
              v-bind:warning="alertType==='warning'" 
              v-bind:error="alertType==='error'" 
              v-model="alertMsg" 
              dismissible>{{ alertMsg }}</v-alert>
            <v-card-row>
              <v-card-title>修改目录</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field 
                  v-model="UpdateTree.Name"
                  ref="UpdateTree_Name" 
                  label="名称" 
                  required
                  :rules="rules.Update.Name"
                  @input="rules.Update.Name = rules0.Update.Name"
                ></v-text-field>
                <v-text-field 
                  v-model="UpdateTree.Description" 
                  label="说明" 
                  class="mt-4"
                ></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="updateTree">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="UpdateTreeDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-layout row justify-center>
        <v-dialog v-model="RemoveConfirmDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>提示</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>你确认要删除参数目录{{ SelectedTree.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeTree">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-data-table
        :headers="headers"
        :items="items"
        hide-actions
        class="trees-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td>{{ props.item.Id }}</td>
          <td><router-link :to="'/env/trees/' + props.item.Id + '/' + encodeURIComponent(props.item.Name)">{{ props.item.Name }}</router-link></td>
          <td>{{ props.item.Description }}</td>
          <td>
            <v-btn outline small icon class="green green--text" @click.native="edit(props.item)" title="修改">
                <v-icon>mode_edit</v-icon>
            </v-btn>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除参数目录">
              <v-icon>close</v-icon>
            </v-btn>
          </td>
        </template>
      </v-data-table>
    </div>
  </v-card>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        headers: [
          { text: 'ID', sortable: false, left: true },
          { text: '名称', sortable: false, left: true },
          { text: '说明', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],

        CreateTreeDlg: false,
        NewTree: { Name: '', Description: '' },

        UpdateTreeDlg: false,
        UpdateTree: { Name: '', Description: '' },

        RemoveConfirmDlg: false,
        SelectedTree: {},

        rules: { Create: {}, Update: {} },

        rules0: {
          Create: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入参数目录名称')
            ]
          },
          Update: {
            Name: [
              v => (v && v.length > 0 ? true : '请输入参数目录名称')
            ]
          }
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

    watch: {
        CreateTreeDlg(v) {
          (v ? ui.showAlertAt('CreateTreeDlg') : ui.showAlertAt())
        }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.EnvTrees().then(data => {
          this.items = data;
        })
      },

      createTree() {
        this.rules.Create = this.rules0.Create;
        this.$nextTick(_ => {
          if (!this.validateForm('NewTree_')) {
            return;
          }

          this.CreateTreeDlg = false;
          api.CreateEnvTree(this.NewTree).then(data => {
            this.init();
          });
        });
      },

      edit(tree) {
        this.UpdateTree.Id = tree.Id;
        this.UpdateTree.Name = tree.Name;
        this.UpdateTree.Description = tree.Description;
        this.UpdateTreeDlg = true;
      },

      updateTree() {
        this.rules.Update = this.rules0.Update;
        this.$nextTick(_ => {
          if (!this.validateForm('UpdateTree_')) {
            return;
          }

          api.UpdateEnvTree(this.UpdateTree).then(data => {
            this.UpdateTreeDlg = false;
            this.init();
          });
        });
      },

      confirmBeforeRemove(tree) {
        this.SelectedTree = tree;
        this.RemoveConfirmDlg = true;
      },

      removeTree() {
        this.RemoveConfirmDlg = false;
        api.RemoveEnvTree(this.SelectedTree.Id).then(data => {
          this.init();
        })
      }
    }
  }
</script>

<style lang="stylus">
.trees-table
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
