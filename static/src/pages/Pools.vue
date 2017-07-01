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
                <v-text-field ref="Name" label="名称" v-model="NewPool.Name" :rules="rules.Name"></v-text-field>
                <v-select
                  :items="DriverList"
                  v-model="NewPool.Driver"
                  ref="Driver"
                  label="驱动类型"
                  dark
                  single-line
                  auto
                  :rules="rules.Driver"
                ></v-select>
                <v-select
                  :items="NetworkList"
                  v-model="NewPool.Network"
                  ref="Network"
                  label="网络类型"
                  dark
                  single-line
                  auto
                  :rules="rules.Network"
                ></v-select>
                <v-text-field ref="EndPoint" label="API地址" v-model="NewPool.EndPoint" :rules="rules.EndPoint"></v-text-field>
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
          <td>{{ props.item.Network }}</td>
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
          { text: '网络类型', sortable: false, left: true },
          { text: '节点', sortable: false },
          { text: 'CPU', sortable: false },
          { text: '内存 (GB)', sortable: false },
          { text: '磁盘 (GB)', sortable: false },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],
        DriverList: [ 'Swarm', 'Kubernetes' ],
        NetworkList: [ 'Flannel', 'VxLAN' ],
        CreatePoolDlg: false,
        NewPool: { Name: '', Driver: '', Network: '', EndPoint: '' },
        RemoveConfirmDlg: false,
        SelectedPool: {},

        rules: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入集群名称')
          ],
          Driver: [
            v => (v && v.length > 0 ? true : '请选择驱动类型')
          ],
          Network: [
            v => (v && v.length > 0 ? true : '请选择网络类型')
          ],
          EndPoint: [
            v => (v && v.length > 0 ? true : '请输入集群API地址')
          ]
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
        for (let f in this.$refs) {
          let e = this.$refs[f];
          console.log(e);
          if (e.errorBucket && e.errorBucket.length > 0) {
            return;
          }
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
