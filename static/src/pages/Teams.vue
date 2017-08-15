<template>
  <v-card>
    <v-card-title>
      团队管理
      <v-spacer></v-spacer>
      <v-layout row justify-center style="margin-right:0;">
        <v-dialog v-model="CreateTeamDlg">
          <v-btn class="primary white--text" slot="activator"><v-icon light>add</v-icon>新增团队</v-btn>
          <v-card>
            <v-alert 
              v-if="alertArea==='CreateTeamDlg'"
              v-bind:success="alertType==='success'" 
              v-bind:info="alertType==='info'" 
              v-bind:warning="alertType==='warning'" 
              v-bind:error="alertType==='error'" 
              v-model="alertDisplay" 
              dismissible>{{ alertMsg }}</v-alert>
            <v-card-row>
              <v-card-title>新增团队</v-card-title>
            </v-card-row>
            <v-card-row>
              <v-card-text>
                <v-text-field 
                  v-model="NewTeam.Name"
                  ref="Name" 
                  label="名称" 
                  required
                  :rules="rules.Name"
                  @input="rules.Name = rules0.Name"
                ></v-text-field>
                <v-text-field 
                  v-model="NewTeam.Description" 
                  label="说明" 
                  class="mt-4"
                ></v-text-field>
              </v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="blue--text darken-1" flat @click.native="createTeam">确认</v-btn>
              <v-btn class="blue--text darken-1" flat @click.native="CreateTeamDlg = false">取消</v-btn>
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
              <v-card-text>你确认要删除团队{{ SelectedTeam.Name }}吗？</v-card-text>
            </v-card-row>
            <v-card-row actions>
              <v-btn class="green--text darken-1" flat="flat" @click.native="removeTeam">确认</v-btn>
              <v-btn class="green--text darken-1" flat="flat" @click.native="RemoveConfirmDlg = false">取消</v-btn>
            </v-card-row>
          </v-card>
        </v-dialog>
      </v-layout>
      <v-data-table
        :headers="headers"
        :items="items"
        hide-actions
        class="teams-table elevation-1"
        no-data-text=""
      >
        <template slot="items" scope="props">
          <td><router-link :to="'/teams/' + props.item.Id">{{ props.item.Name }}</router-link></td>
          <td>{{ props.item.Leader.Name }}</td>
          <td>{{ props.item.Description }}</td>
          <td>{{ props.item.CreatedTime | formatDate }}</td>
          <td>
            <v-btn outline small icon class="orange orange--text" @click.native="confirmBeforeRemove(props.item)" title="删除团队">
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
          { text: '名称', sortable: false, left: true },
          { text: '主管', sortable: false, left: true },
          { text: '说明', sortable: false, left: true },
          { text: '创建时间', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],
        items: [],

        CreateTeamDlg: false,
        NewTeam: { Name: '', Description: '' },

        RemoveConfirmDlg: false,
        SelectedTeam: {},

        rules: {},

        rules0: {
          Name: [
            v => (v && v.length > 0 ? true : '请输入团队名称')
          ]
        }
      }
    },

    computed: {
      ...mapGetters([
          'alertArea',
          'alertType',
          'alertMsg'
      ]),

      alertDisplay: {
        get() {
          return this.$store.getters.alertArea != null;
        },
        set(v) {
          this.$store.dispatch('alertArea', null);
        }
      }
    },

    watch: {
        CreateTeamDlg(v) {
          (v ? ui.showAlertAt('CreateTeamDlg') : ui.showAlertAt())
        }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Teams().then(data => {
          this.items = data;
        })
      },

      createTeam() {
        this.rules = this.rules0;
        this.$nextTick(_ => {
          if (!this.validateForm()) {
            return;
          }

          this.CreateTeamDlg = false;
          api.CreateTeam(this.NewTeam).then(data => {
            this.init();
          });
        });
      },

      confirmBeforeRemove(team) {
        this.SelectedTeam = team;
        this.RemoveConfirmDlg = true;
      },

      removeTeam() {
        this.RemoveConfirmDlg = false;
        api.RemoveTeam(this.SelectedTeam.Id).then(data => {
          this.init();
        })
      }
    }
  }
</script>

<style lang="stylus">
.teams-table
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
