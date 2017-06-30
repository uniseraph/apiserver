<template>
  <v-layout row>
  <v-flex xs4>
    <v-card>
      <v-card-title>
        参数目录
        <v-spacer></v-spacer>
      </v-card-title>
      <div class="pa-2">
        <tree ref="tree" :options="treeOptions" :treeData="treeData" @node-click="nodeClicked" />
      </div>
    </v-card>
  </v-flex>
  <v-flex xs8>
    <v-card>
      <v-data-table
        v-bind:headers="headers"
        v-bind:items="items"
        hide-actions
        class="elevation-1"
        no-data-text=""
      >
        <template slot="headers" scope="props">
          <span>
            {{ props.item.text }}
          </span>
        </template>
        <template slot="items" scope="props">
          <td>{{ props.item.Id }}</td>
          <td>{{ props.item.Name }}</td>
          <td>{{ props.item.Value }}</td>
          <td>{{ props.item.Description }}</td>
        </template>
      </v-data-table>
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
          { text: '描述', sortable: false, left: true }
        ],
        items: []
      }
    },

    mounted() {
      this.init();
    },

    methods: {
      init() {
        api.EnvDirs().then(data => {
          let treeData = [{
            id: "0",
            label: '全部',
            open: true,
            visible: true,
            checked: false,
            children: conv2TreeData(data)
          }];

          this.treeData = treeData;
        })
      },

      nodeClicked(node) {
        api.EnvList({ DirId: node.id }).then(data => {
          this.items = data;
        })
      }
    },

    components: {
      Tree
    }
  }

  function conv2TreeData(list) {
    let arr = [];
    for (let e of list) {
      let a = {
        id: e.Id,
        label: e.Name,
        parentId: e.ParentId ? e.ParentId : "0",
        open: false,
        visible: true,
        checked: false
      };
      arr.push(a);
      if (e.Children) {
        a.children = conv2TreeData(e.Children);
      }
    }

    return arr;
  }
</script>

<style lang="stylus">

</style>
