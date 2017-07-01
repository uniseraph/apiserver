<template>
  <ul>
    <li v-for='item in nodeData' v-show="item.visible" :class="{ 'root': item.parentId === undefined, 'leaf': item.parentId && (!item.children || item.children.length == 0) }">
      <i v-if="item.children && item.children.length > 0"  @click.stop='handleNodeExpand(item)' :class="[item.open? 'tree-open':'tree-close','icon']">
        </i>
     
      <span class="node-label" @click="handleNode(item)" :class="{'node-selected':(item.checked && !options.showCheckbox) || item.searched }">{{item.label}}</span>

      <tree-node v-if="item.children && item.children.length > 0" :options="options" @handleCheckedChange="handleCheckedChange" v-show='item.open'
        :nodeData="item.children"></tree-node>
    </li>
  </ul>
</template>
<script>
export default {
  name: 'treeNode',
  props: [ 'nodeData', 'options' ],

  data() {
    return {

    }
  },

  created() {
    const parent = this.$parent
    if (parent.isTree) {
      this.tree = parent
    } else {
      this.tree = parent.tree
    }
  },

  computed: {
    inputWidth() {
      if (this.checkFirfox()) {
        return 14
      }

      return 13
    }
  },

  methods: {
    checkFirfox(){
      let u = navigator.userAgent
      if (u.indexOf('Gecko') > -1 && u.indexOf('KHTML') == -1) {
        return true
      }

      return false
    },

    handleNodeExpand(node) {
      node.open = !node.open
    },

    handleCheckedChange(node) {
      this.$emit('handleCheckedChange', node)
    },

    handleNode(node) {
      if (this.tree.store.last) {
        if (this.tree.store.last.id === node.id) {
          //this.tree.store.last.checked = !this.tree.store.last.checked
        } else {
          this.tree.store.last.checked = false
          node.checked = true
          this.tree.store.last = node
        }
      } else {
        node.checked = true
        this.tree.store.last = node
      }

      this.tree.$emit('node-click', node)
    }
  }
}
</script>
<style scoped>
  li:hover {
    cursor: pointer;
  }

  .icon{
  display: inline-block;
  margin-right: 10px;
  vertical-align: middle;
   }
   
  .halo-tree {
    font-size: 14px;
    min-height: 20px;
    -webkit-border-radius: 4px;
    -moz-border-radius: 4px;
    border-radius: 4px;
  }

  .halo-tree li {
    margin: 0;
    padding: 5px 5px 5px 0;
    position: relative;
    list-style: none;
  }
  
  .halo-tree li > span,
  .halo-tree li > i,
  .halo-tree li > a {
    line-height: 20px;
    vertical-align: middle;
  }
  
  .halo-tree .node-label.node-selected {
    background: #64b5f6;
    color: #FFF;
  }
  
  .halo-tree li:after,
  .halo-tree li:before {
    content: '';
    left: -13px;
    position: absolute;
    right: auto;
    border-width: 1px
  }
  
  .halo-tree li:before {
    border-left: 1px dashed #999;
    bottom: 50px;
    height: 100%;
    top: -8px;
    width: 1px;
  }

  .halo-tree li.root:before {
    border-left: none;
  }
  
  .halo-tree li:after {
    border-top: 1px dashed #999;
    height: 20px;
    top: 17px;
    width: 12px
  }
  
  .halo-tree li.root:after {
    border-top: none;
  }

  .halo-tree li span {
    display: inline-block;
    padding: 3px 3px;
    text-decoration: none;
    border-radius: 3px;
  }
  
  .halo-tree li:last-child::before {
    height: 26px
  }
  
  .halo-tree > ul {
    padding-left: 0
  }
  
  .halo-tree ul ul {
    padding-left: 17px;
    padding-top: 10px;
  }
  
  .halo-tree li.leaf {
    padding-left: 21px;
  }
  
  .halo-tree li.leaf:after {
    content: '';
    left: -13px;
    position: absolute;
    right: auto;
    border-width: 1px;
    border-top: 1px dashed #999;
    height: 20px;
    top: 17px;
    width: 24px;
  }
  
  .check {
    display: inline-block;
    position: relative;
    top: 4px;
  }
  
  .halo-tree .icon {
    margin-right: 0;
  }

  .tree-close{
  width:16px;
  height:16px;
  background-image: url("/public/tree/close.gif");
}

.tree-open{
  width:16px;
  height:16px;
  background-image: url("/public/tree/open.gif");
}
.search{
  width:14px;
  height:14px;
  background-image: url("/public/tree/search.png");
}
/*.check.notAllNodes{
  -webkit-appearance: none;
  -moz-appearance: none;
  -ms-appearance: none;
  width: 13px;
}*/
.inputCheck{
  display: inline-block;
  position: relative;
}
.inputCheck.notAllNodes:before{
  content: "";
  display: inline-block;
  position: absolute;
  width: 100%;
  height: 100%;
  z-index: 10;
  top: 50%;
  left: 50%;
  transform: translate3d(-30%,-5%,0);
  /*background-image: url("/static/images/half.png");*/
  background-image: url("/public/tree/half.jpg");
}
</style>