export default {
  install(Vue, options) {
    Vue.prototype.isEmail = function(s) {
    	return /^((([a-z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+(\.([a-z]|\d|[!#\$%&'\*\+\-\/=\?\^_`{\|}~]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])+)*)|((\x22)((((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(([\x01-\x08\x0b\x0c\x0e-\x1f\x7f]|\x21|[\x23-\x5b]|[\x5d-\x7e]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(\\([\x01-\x09\x0b\x0c\x0d-\x7f]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF]))))*(((\x20|\x09)*(\x0d\x0a))?(\x20|\x09)+)?(\x22)))@((([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-z]|\d|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.)+(([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])|(([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])([a-z]|\d|-|\.|_|~|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])*([a-z]|[\u00A0-\uD7FF\uF900-\uFDCF\uFDF0-\uFFEF])))\.?$/i.test(s.trim());
    },

    Vue.prototype.moduleName = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      return Vue.prototype.constants.MODULE_MAP[id].Name;
    },

    Vue.prototype.operationName = function(moduleId, id) {
      if (!id || id.length == 0) {
        return '';
      }

      let list = Vue.prototype.constants.OPERATION_MAP[moduleId];
      for (let e of list) {
        if (id == e.Id) {
          return e.Name;
        }
      }

      return '';
    },

    Vue.prototype.sshOpName = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let list = Vue.prototype.constants.SSH_OP_LIST;
      for (let e of list) {
        if (id == e.Id) {
          return e.Name;
        }
      }

      return '';
    },

    Vue.prototype.applicationStatus = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.APPLICATION_STATUS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.applicationClass = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.APPLICATION_CLASS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.containerStatus = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.CONTAINER_STATUS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },

    Vue.prototype.containerClass = function(id) {
      if (!id || id.length == 0) {
        return '';
      }

      let m = Vue.prototype.constants.CONTAINER_CLASS_MAP;
      let s = m[id];
      return s ? s : m["*"];
    },
    
    Vue.prototype.validateForm = function(refPrefix) {
      let result = true;
    	Object.keys(this.$refs).forEach(k => {
        if (!refPrefix || k.indexOf(refPrefix) == 0) {
          let e = this.$refs[k];
          if (Array.isArray(e)) {
            e = e[0];
          }

          if (e && e.validate) {
            e.validate();
            if (e.errorBucket && e.errorBucket.length > 0) {
              result = false;
            }
          }
        }
      });

      return result;
    }
  }
}