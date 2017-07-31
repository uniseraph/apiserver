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
          <v-flex xs4>
            <v-subheader>节点个数：{{ Summary.Nodes }}</v-subheader>
          </v-flex>
          <v-flex xs4>
            <v-subheader>应用个数：{{ Summary.Applications }}</v-subheader>
          </v-flex>
          <v-flex xs4>
            <v-subheader>容器个数：{{ Summary.Containers }}</v-subheader>
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
          
        </v-layout>
      </v-container>
    </div>
  </v-card>
</div>
</template>

<script>
  import api from '../api/api'
  import * as ui from '../util/ui'

  export default {
    data() {
      return {
        PoolId: '',
        PoolList: [],
        TimeList: [ 
          { Label: '最近7天发布统计', Value: 7 },
          { Label: '最近15天发布统计', Value: 15 },
          { Label: '最近30天发布统计', Value: 30 }
        ],
        StartTime: 7,
        Summary: {},
        Trend: {}
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.Pools().then(data => {
          this.PoolList = [{ Id: '', Name: '所有集群' }].concat(data);
        })

        this.stat();
      },

      stat() {
        let d = new Date(new Date().getTime() - 1000 * 60 * 60 * 24 * this.StartTime);
        let st = d.getFullYear() + "-" + (d.getMonth() + 1) + "-" + d.getDate();

        api.Stat({
          PoolId: this.PoolId,
          StartTime: st
        }).then(data => {
          this.Summary = data.Summary;
          this.Trend = data.Trend;
        })
      }
    }
  }
</script>

<style lang="stylus">

</style>
