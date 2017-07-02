<template>
<div class="halo-tree">
 <tree-node :nodeData="treeData.nodeData" :options="options" @handleCheckedChange="handleCheckedChange"></tree-node>
</div>
</template>
<script>
  import TreeNode from './tree-node.vue'
  import TreeStore from './tree-store'
  export default {
    name: 'tree',
    props: [ 'treeData', 'options' ],

    data() {
      return {
        search: null
      }
    },

    created() {
      this.isTree = true
      this.refresh()
    },

    watch: {
      search(val) {
        this.store.filterNodes(val, this.options.search)
      },

      treeData(val) {
        this.refresh()
      }
    },

    methods: {
      handleCheckedChange(node) {
        if (this.options.halfCheckedStatus) {
          this.store.changeCheckHalfStatus(node)
        } else {
          this.store.changeCheckStatus(node)
        }

        this.$emit('handleCheckedChange', node)
      },

      getSelectedNodes() {
        const allnodes = this.store.datas
        let selectedNodes = []
        for (let [, node] of allnodes) {
          if (node.checked) {
            selectedNodes.push(node)
          }
        }

        return selectedNodes
      },

      getSelectedNodeIds() {
        const allnodes = this.store.datas
        let selectedNodeIds = []
        for (let [, node] of allnodes) {
          if (node.checked) {
            selectedNodeIds.push(node.id)
          }
        }

        return selectedNodeIds
      },

      getState() {
        let state = {};
        const traverseNodes = (root) => {
            for (let node of root) {
                let s = {
                  open: node.open,
                  visible: node.visible,
                  checked: false // node.checked 先只支持单选
                };

                state[node.id] = s;
                if (node.children && node.children.length > 0) traverseNodes(node.children)
            }
        };

        traverseNodes(this.treeData.nodeData);
        return state;
      },

      createTreeData(nodeData, state, currNodeId) {
        let parentNode = null;
        const traverseNodes = (root) => {
            for (let node of root) {
                let s = state[node.id];
                if (s) {
                  node.open = s.open;
                  node.visible = s.visible;
                  node.checked = s.checked;
                }

                if (currNodeId && node.id == currNodeId) {
                  if (parentNode) {
                    parentNode.open = true;
                  }

                  node.checked = true;
                }

                if (node.children && node.children.length > 0) {
                  parentNode = node;
                  traverseNodes(node.children)
                }
            }
        };

        traverseNodes(nodeData);

        return { nodeData: nodeData, currentNodeId: currNodeId };
      },

      getNodeById(id) {
        let lastNode = null;
        const traverseNodes = (root) => {
            for (let node of root) {
                if (id && node.id == id) {
                  lastNode = node;
                  return;
                }

                if (node.children && node.children.length > 0) traverseNodes(node.children)
            }
        };

        traverseNodes(this.treeData.nodeData);
        return lastNode;
      },

      refresh() {
        let lastNode = this.getNodeById(this.treeData.currentNodeId);
        this.store = new TreeStore({ root: this.treeData.nodeData, last: lastNode });
      }
    },

    components: { TreeNode }
  }
</script>
<style scoped>
  *{
    font-size: 13px;
    font-family: '\5FAE\8F6F\96C5\9ED1'
  }
  .input{
    width:100%;
    position: relative;
  }
  .input span {
    position: absolute;
    top:7px;
    right:5px;
  }
  .input input{
    display: inline-block;
    box-sizing: border-box;
    width:100%;
    border-radius: 5px;
    height:25px;
    margin-top: 2px;
  }
  .input input:focus {
      border:none;
  }
  .search{
  width:14px;
  height:14px;
  background-image: url("/public/tree/search.png");
}
</style>

