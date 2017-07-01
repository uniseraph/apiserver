<template>
  <v-card>
    <v-card-title>
      集群列表
      <v-spacer></v-spacer>
      <v-layout row justify-center style="margin-right:0;">
        <v-dialog v-model="CreatePoolDlg">
          <v-btn class="primary white--text" slot="activator"><v-icon light>add</v-icon>新增集群</v-btn>
          <v-card>
            <v-card-row>
              <v-card-title>新增集群</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field 
                  v-model="NewPool.Name" 
                  ref="all_Name" 
                  label="名称" 
                  required 
                  persistent-hint 
                  :rules="rules.Name"
                ></v-text-field>
                <v-select
                  :items="DriverList"
                  v-model="NewPool.Driver"
                  ref="all_Driver"
                  label="驱动类型"
                  dark
                  required
                  :rules="rules.Driver"
                  class="mt-4"
                ></v-select>
                <v-select
                  v-if="NewPool.Driver == 'swarm'"
                  :items="SwarmVersionList"
                  v-model="NewPool.DriverOpts.Version"
                  ref="swarm_Version"
                  label="驱动版本"
                  dark
                  required
                  :rules="rules.DriverOpts.swarm.Version"
                  class="mt-4"
                ></v-select>
                <v-text-field 
                  v-if="NewPool.Driver == 'swarm'"
                  v-model="NewPool.DriverOpts.EndPoint" 
                  ref="swarm_EndPoint" 
                  label="API地址" 
                  required 
                  :rules="rules.DriverOpts.swarm.EndPoint"
                  class="mt-4"
                ></v-text-field>
                <v-select
                  v-if="NewPool.Driver == 'swarm'"
                  :items="SwarmAPIVersionList"
                  v-model="NewPool.DriverOpts.APIVersion"
                  ref="swarm_APIVersion"
                  label="驱动版本"
                  dark
                  required
                  :rules="rules.DriverOpts.swarm.APIVersion"
                  class="mt-4"
                ></v-select>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="createPool">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="CreatePoolDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
    </v-card-title>
    <div>
      <v-layout row justify-center>
        <v-dialog v-model="RemoveConfirmDlg" persistent>
          <v-card>
            <v-card-row>
              <v-card-title>提示</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>你确认要删除集群{{ SelectedPool.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removePool">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-data-table
        :headers="headers"
        :items="items"
        hide-actions
        class="pools-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td>{{ props.item.Id }}</td>
          <td><router-link :to="'/pool/' + props.item.Id + '/detail'">{{ props.item.Name }}</router-link></td>
          <td>{{ props.item.Driver }}</td>
          <td class="text-xs-right">{{ props.item.Nodes }}</td>
          <td class="text-xs-right">
            {{ props.item.Cpus }}
          </td>
          <td class="text-xs-right">
            {{ props.item.Memories }}
          </td>
          <td class="text-xs-right">
            {{ props.item.Disks }}
          </td>
          <td>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除集群">
              <v-icon>close</v-icon>
            </v-btn>
          </td>
        </template>
      </v-data-table>
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
          { text: 'ID', sortable: false, left: true },
          { text: '名称', sortable: false, left: true },
          { text: '驱动类型', sortable: false, left: true },
          { text: '节点', sortable: false },
          { text: 'CPU', sortable: false },
          { text: '内存 (GB)', sortable: false },
          { text: '磁盘 (GB)', sortable: false },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],

        DriverList: [ 'swarm' ],
        SwarmVersionList: [ 'v1.0' ],
        SwarmAPIVersionList: [ 'v1.23' ],

        CreatePoolDlg: false,
        NewPool: { Name: '', Driver: 'swarm', DriverOpts: { Version: 'v1.0', EndPoint: '', APIVersion: 'v1.23' } },

        RemoveConfirmDlg: false,
        SelectedPool: {},

        rules: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入集群名称')
          ],
          Driver: [
            v => (v && v.length > 0 ? true : '请选择驱动类型')
          ],
          DriverOpts: {
            swarm: {
              Version: [
                v => (v && v.length > 0 ? true : '请选择集群驱动版本')
              ],
              EndPoint: [
                v => (v && v.length > 0 ? true : '请输入集群API地址')
              ],
              APIVersion: [
                v => (v && v.length > 0 ? true : '请选择集群API版本')
              ]
            }
          }
        }
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Pools().then(data => {
          this.items = data;
        })
      },

      createPool() {
        if (!this.validateForm('all_') || !this.validateForm(this.Driver + '_')) {
          return;
        }

        this.CreatePoolDlg = false;
        api.CreatePool(this.NewPool).then(data => {
          this.init();
        })
      },

      confirmBeforeRemove(pool) {
        this.SelectedPool = pool;
        this.RemoveConfirmDlg = true;
      },

      removePool() {
        this.RemoveConfirmDlg = false;
        api.RemovePool({ Id: this.SelectedPool.Id }).then(data => {
          this.init();
        })
      }
    }
  }
</script>

<style lang="stylus">
.pools-table
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
