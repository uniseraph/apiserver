<template>
<div>
  <v-card>
    <v-card-title>
      集群概况
      <v-spacer></v-spacer>
      <v-select
          :items="PoolList"
          item-text="Name"
          item-value="Id"
          v-model="PoolId"
          dark
          @input="stat"></v-select>
      </v-layout>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs3 class="text-xs-center">
            <v-icon class="blue--text">device_hub</v-icon>
            &nbsp;节点个数：{{ Summary.Nodes }}
          </v-flex>
          <v-flex xs1>
          </v-flex>
          <v-flex xs3 class="text-xs-center">
            <v-icon class="green--text">brightness_auto</v-icon>
            &nbsp;应用个数：{{ Summary.Applications }}
          </v-flex>
          <v-flex xs1>
          </v-flex>
          <v-flex xs3 class="text-xs-center">
            <v-icon class="cyan--text">directions_boat</v-icon>
            &nbsp;容器个数：{{ Summary.Containers }}
          </v-flex>
          <v-flex xs1>
          </v-flex>
          <v-flex xs3 mt-4>
            <pie-chart :chart-data="CPUUsageData" :options="CPUUsageOptions"></pie-chart>
          </v-flex>
          <v-flex xs1 mt-4>
          </v-flex>
          <v-flex xs3 mt-4>
            <pie-chart :chart-data="MemoryUsageData" :options="MemoryUsageOptions"></pie-chart>
          </v-flex>
          <v-flex xs1 mt-4>
          </v-flex>
          <v-flex xs3 mt-4>
            <pie-chart :chart-data="DiskUsageData" :options="DiskUsageOptions"></pie-chart>
          </v-flex>
          <v-flex xs1 mt-4>
          </v-flex>
        </v-layout>
      </v-container>
    </div>
  </v-card>
  <v-card class="mt-4">
    <v-card-title>
      应用发布统计
      <v-spacer></v-spacer>
      <v-select
          :items="TimeList"
          item-text="Label"
          item-value="Value"
          v-model="StartTime"
          dark
          @input="stat"></v-select>
      </v-layout>
    </v-card-title>
    <div>
      <v-container fluid>
        <v-layout row wrap>
          <v-flex xs12>
            <bar-chart :chart-data="VersionData" :options="VersionOptions" :height="100"></bar-chart>
          </v-flex>
          <v-flex xs6 mt-4>
            <v-subheader>升级次数最多的应用</v-subheader>
            <v-data-table
              :headers="headers_upgrade"
              :items="Trend.MostUpgradeApplications"
              hide-actions
              class="app-table elevation-1"
              no-data-text=""
            >
              <template slot="items" scope="props">
                <td><router-link :to="'/applications/' + props.item.Id + '/' + encodeURIComponent(PoolMap[PoolId])">{{ props.item.Title }}</router-link></td>
                <td>{{ props.item.Name }}</td>
                <td>{{ props.item.Version }}</td>
                <td class="text-xs-right">{{ props.item.Count }}</td>
              </template>
            </v-data-table>
          </v-flex>
          <v-flex xs6 mt-4>
            <v-subheader>回滚次数最多的应用</v-subheader>
            <v-data-table
              :headers="headers_rollback"
              :items="Trend.MostRollbackApplications"
              hide-actions
              class="app-table elevation-1"
              no-data-text=""
            >
              <template slot="items" scope="props">
                <td><router-link :to="'/applications/' + props.item.Id + '/' + encodeURIComponent(PoolMap[PoolId])">{{ props.item.Title }}</router-link></td>
                <td>{{ props.item.Name }}</td>
                <td>{{ props.item.Version }}</td>
                <td class="text-xs-right">{{ props.item.Count }}</td>
              </template>
            </v-data-table>
          </v-flex>
        </v-layout>
      </v-container>
    </div>
  </v-card>
</div>
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'
  import PieChart from '../components/chart/PieChart'
  import BarChart from '../components/chart/BarChart'

  export default {
    components: {
      PieChart,
      BarChart
    },

    data() {
      return {
        headers_upgrade: [
          { text: '应用名称', sortable: false, left: true },
          { text: '应用ID', sortable: false, left: true },
          { text: '当前版本', sortable: false, left: true },
          { text: '升级次数', sortable: false, left: true }
        ],
        headers_rollback: [
          { text: '应用名称', sortable: false, left: true },
          { text: '应用ID', sortable: false, left: true },
          { text: '当前版本', sortable: false, left: true },
          { text: '回滚次数', sortable: false, left: true }
        ],

        PoolId: this.$route.query ? (this.$route.query.PoolId || '') : '', 
        StartTime: this.$route.query ? parseInt(this.$route.query.StartTime || '7') : 7, 

        PoolList: [],
        PoolMap: {},
        TimeList: [ 
          { Label: '最近7天发布统计', Value: 7 },
          { Label: '最近15天发布统计', Value: 15 },
          { Label: '最近30天发布统计', Value: 30 }
        ],

        Summary: {},
        Trend: {},

        CPUUsageData: null,
        CPUUsageOptions: {
          title: {
            display: true,
            text: "CPU分配情况"
          },
          tooltips: {
            callbacks: {
              label: function(a, data) {
                return data.labels[a.index];
              }
            }
          }
        },

        MemoryUsageData: null,
        MemoryUsageOptions: {
          title: {
            display: true,
            text: "内存分配情况 (GB)"
          },
          tooltips: {
            callbacks: {
              label: function(a, data) {
                return data.labels[a.index];
              }
            }
          }
        },

        DiskUsageData: null,
        DiskUsageOptions: {
          title: {
            display: true,
            text: "硬盘分配情况 (GB)"
          },
          tooltips: {
            callbacks: {
              label: function(a, data) {
                return data.labels[a.index];
              }
            }
          }
        },

        VersionData: null,
        VersionOptions: {
          scales: {
            yAxes: [{
              ticks: {
                beginAtZero: true
              },
              gridLines: {
                display: true
              }
            }],
            xAxes: [{
              gridLines: {
                display: true
              },
              categoryPercentage: 0.9,
              barPercentage: 0.8
            }]
          }
        }
      }
    },

    created() {
      this.init();
    },

    methods: {
      init() {
        api.Pools().then(data => {
          this.PoolList = data;

          this.PoolMap = {};
          for (let p of data) {
            this.PoolMap[p.Id] = p.Name;
          }

          if (data.length > 0) {
            if (this.PoolId == '') {
              this.PoolId = data[0].Id;
            } else {
              this.stat();
            }
          }
        })
      },

      stat() {
        let params = {
          PoolId: this.PoolId,
          StartTime: this.StartTime
        };

        this.$router.replace({
          name: this.$route.name,
          params: this.$route.params,
          query: params
        });

        let d = new Date(new Date().getTime() - 1000 * 60 * 60 * 24 * this.StartTime);
        let mon = d.getMonth() + 1;
        let dat = d.getDate();
        let st = d.getFullYear() + "-" + (mon > 9 ? mon : '0' + mon) + "-" + (dat > 9 ? dat : '0' + dat);

        api.Stat({
          PoolId: this.PoolId,
          StartTime: st
        }).then(data => {
          let s = data.Summary;
          this.CPUUsageData = {
            datasets: [{
              data: [ s.CPUsUsed, s.CPUs - s.CPUsUsed ],
              backgroundColor: [ "#FF6384", "#4BC0C0" ]
            }],
            labels: [ '独占' + s.CPUsUsed, '共享' + (s.CPUs - s.CPUsUsed) ]
          };
          let mu = parseInt(s.MemoryUsed / 1024 / 1024 / 1024);
          let muu = parseInt((s.Memory - s.MemoryUsed) / 1024 / 1024 / 1024);
          this.MemoryUsageData = {
            datasets: [{
              data: [ mu, muu ],
              backgroundColor: [ "#FF6384", "#4BC0C0" ]
            }],
            labels: [ '已分配' + mu, '剩余' + muu ]
          };
          let du = parseInt(s.DiskUsed / 1024 / 1024 / 1024);
          let duu = parseInt((s.Disk - s.DiskUsed) / 1024 / 1024 / 1024);
          this.DiskUsageData = {
            datasets: [{
              data: [ du, duu ],
              backgroundColor: [ "#FF6384", "#4BC0C0" ]
            }],
            labels: [ '已分配' + du, '剩余' + duu ]
          };

          let t = data.Trend;
          let labels = [];
          let creates = [];
          let upgrades = [];
          let rollbacks = [];
          for (let r of t.Creates) {
            labels.push(r.Day);
            creates.push(r.Count);
          }
          for (let r of t.Upgrades) {
            upgrades.push(r.Count);
          }
          for (let r of t.Rollbacks) {
            rollbacks.push(r.Count);
          }

          this.VersionData = {
            datasets: [{
              label: '新增应用个数',
              data: creates,
              backgroundColor: "rgba(54, 162, 235, 0.2)",
              borderColor: "rgb(54, 162, 235)",
              borderWidth: 1
            }, {
              label: '应用升级次数',
              data: upgrades,
              backgroundColor: "rgba(75, 192, 192, 0.2)",
              borderColor: "rgba(75, 192, 192, 1)",
              borderWidth: 1
            }, {
              label: '应用回滚次数',
              data: rollbacks,
              backgroundColor: "rgba(255, 99, 132, 0.2)",
              borderColor: "rgba(255, 99, 104, 1)",
              borderWidth: 1
            }],
            labels: labels
          },

          this.Summary = data.Summary;
          this.Trend = data.Trend;
        })
      }
    }
  }
</script>

<style lang="stylus">
</style>
