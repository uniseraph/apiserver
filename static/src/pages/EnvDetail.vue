<template>
  <v-layout column>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          <i class="material-icons ico_back" @click="goback">keyboard_arrow_left</i>
          &nbsp;&nbsp;参数目录&nbsp;&nbsp;/&nbsp;&nbsp;{{ Name }}
          <v-spacer></v-spacer>
        </v-card-title>
        <div>
          <v-container fluid>
            <v-layout row wrap>
              <v-flex xs2>
                <v-subheader>参数名<span class="required-star">*</span></v-subheader>
              </v-flex>
              <v-flex xs3>
                <v-text-field
                  v-model="Name"
                  ref="Env_Name"
                  required
                  :rules="rules.Env.Name"
                  @input="rules.Env.Name = rules0.Env.Name"
                ></v-text-field>
              </v-flex>
              <v-flex xs2>
              </v-flex>
              <v-flex xs2>
                <v-checkbox label="敏感数据" v-model="Mask" dark></v-checkbox>
              </v-flex>
              <v-flex xs3>
              </v-flex>
              <v-flex xs2>
                <v-subheader>默认值<span class="required-star">*</span></v-subheader>
              </v-flex xs2>
              <v-flex xs7>
                <v-text-field
                  v-model="Value"
                  ref="Env_Value"
                  required
                  :rules="rules.Env.Value"
                  @input="rules.Env.Value = rules0.Env.Value"
                  class="completer-field"
                  rel="Env_Value"
                ></v-text-field>
              </v-flex>
              <v-flex xs3>
              </v-flex>
              <v-flex xs2>
                <v-subheader>说明</v-subheader>
              </v-flex>
              <v-flex xs7>
                <v-text-field
                  v-model="Description"
                  ref="Name"
                ></v-text-field>
              </v-flex>
              <v-flex xs3>
              </v-flex>
              <v-flex xs12 mt-4 class="text-xs-center">
                <v-btn class="orange darken-2 white--text" @click.native="saveEnvValue">
                  <v-icon light left>save</v-icon>保存
                </v-btn>            
              </v-flex>
            </v-layout>
          </v-container>
        </div>
      </v-card>
    </v-flex>
    <v-flex xs12 mt-4>
      <v-alert 
            v-if="alertArea==='PoolValueAlert'"
            v-bind:success="alertType==='success'" 
            v-bind:info="alertType==='info'" 
            v-bind:warning="alertType==='warning'" 
            v-bind:error="alertType==='error'" 
            v-model="alertDisplay" 
            dismissible>{{ alertMsg }}</v-alert>
    </v-flex>
    <v-flex xs12>
      <v-card>
        <v-card-title>
          各集群当前参数值
          <v-spacer></v-spacer>
        </v-card-title>
        <div>
          <v-data-table
            :headers="headers"
            :items="Values"
            hide-actions
            class="elevation-1"
            no-data-text=""
          >
            <template slot="items" scope="props">
              <td>{{ props.item.PoolName }}</td>
              <td>
                <v-text-field
                  v-model="props.item.Value"
                  :ref="'Pool_' + props.item.PoolId"
                  required
                  :rules="rules.Pool.Values"
                  @input="rules.Pool.Values = rules0.Pool.Values"
                  class="completer-field"
                  :rel="'Pool_' + props.item.PoolId"
                ></v-text-field>
              </td>
              <td>
                <v-btn outline small icon class="orange orange--text" @click.native="savePoolValue(props.item)" title="保存">
                  <v-icon>save</v-icon>
                </v-btn>
              </td>
            </template>
          </v-data-table>
        </div>
      </v-card>
    </v-flex>
  </v-layout>
</template>

<script>
  import store, { mapGetters } from 'vuex'
  import api from '../api/api'
  import jQuery from 'jquery'
  import caret from '../caret'
  import completer from '../completer'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        headers: [
          { text: '集群名称', sortable: false, left: true },
          { text: '当前值', sortable: false, left: true },
          { text: '操作', sortable: false, left: true }
        ],

        Id: this.$route.params.id,
        TreeId: this.$route.params.treeId,
        Name: '',
        Mask: false,
        Value: '',
        Description: '',
        Values: [],

        rules: { Env: {}, Pool: {} },

        rules0: {
          Env: {
            Name: [
                v => (v && v.length > 0 ? true : '请输入参数名')
            ],

            Value: [
              v => (v && v.length > 0 ? true : '请输入默认值')
            ]
          },

          Pool: {
            Values: [
              v => (v && v.length > 0 ? true : '请输入集群参数值')
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

    mounted() {
      this.init();
    },

    destroyed() {
      ui.showAlertAt();
    },

    methods: {
      init() {
        api.EnvValue(this.Id).then(data => {
          this.Id = data.Id;
          this.Name = data.Name;
          this.Mask = data.Mask;
          this.Value = data.Value;
          this.Description = data.Description;
          this.Values = data.Values;

          this.initCompleters();
        })
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

      saveEnvValue() {
        ui.showAlertAt();

        this.rules.Env = this.rules0.Env;
        this.$nextTick(_ => {
          if (!this.validateForm('Env_')) {
            ui.alert('请正确填写参数信息');
            return;
          }

          api.UpdateEnvValue({
            Id: this.Id,
            Name: this.Name,
            Mask: this.Mask,
            Value: this.Value,
            Description: this.Description
          }).then(data => {
            ui.alert('参数信息修改成功', 'success');
          });
        });
      },

      savePoolValue(item) {
        ui.showAlertAt('PoolValueAlert');

        this.rules.Pool = this.rules0.Pool;
        this.$nextTick(_ => {
          if (!this.validateForm('Pool_')) {
            return;
          }

          api.UpdatePoolValues(this.Id, [{
            PoolId: item.PoolId,
            Value: item.Value
          }]).then(data => {
            ui.alert('集群参数值修改成功', 'success');
          });
        });
      }
    }
  }

  function filterArray(arr1, arr2, p) {
    let m = array2Map(arr2, p);
    let r = [];
    for (let e of arr1) {
      if (!m.has(e[p])) {
        r.push(e);
      }
    }

    return r;
  }

  function array2Map(arr, p) {
    let m = new Map();
    for (let e of arr) {
      m.set(e[p], e);
    }

    return m;
  }
</script>

<style lang="stylus">

</style>
