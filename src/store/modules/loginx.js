import VueCookie from 'vue-cookie'
import Vue from 'vue'
import VueResource from 'vue-resource'
Vue.use(VueResource)
const loginx = {
  state: {
    LoginError: 'Delete Index Success.',
    LoginFlag: false,
    Action: false,
    list: [
      {
        'uuid': 'uuiidididididididi',
        'localaddr': '127.0.0.1:9200',
        'remoteaddr': '10.187.102.8:9200',
        'middleaddr': '10.187.102.8:22',
        'message': 'Starting SSH Tunnel on 127.0.0.1:9200...',
        'ststus': true,
        'sshPwd': 'tianyan',
        'sshUser': 'root',
        'sshpubkey': ''
      }
    ],
    ListErrorMsg: '',
    AddServerErrorMsg: ''
  },
  mutations: {
    SET_LOGIN_ERROR: (state, text) => {
      state.LoginError = text
    },
    SET_LOGIN_FLAG: (state, bool) => {
      state.LoginFlag = bool
    },
    SET_ACTION: (state, bool) => {
      state.Action = bool
    },
    SET_LIST: (state, list) => {
      state.list = list
    },
    SET_LIST_ERROR_MSG: (state, msg) => {
      state.ListErrorMsg = msg
    },
    SET_ADD_SERVER_ERROR_MSG: (state, msg) => {
      state.AddServerErrorMsg = msg
    }
  },
  actions: {
    CookieSetUserInfo ({ commit }, loginForm) {
      Vue.http.post(loginForm.url, loginForm)
      .then(
        response => {
          if (response.body.result === 0) {
            commit('SET_LOGIN_ERROR', '')
            commit('SET_LOGIN_FLAG', true)
            commit('SET_ACTION', !loginForm.action)
            VueCookie.set('SSHTunnelUserInfo', loginForm, { expires: '1D' })
            // window.location.reload()
          } else {
            commit('SET_LOGIN_ERROR', new Date().toString() + response.body.errormsg)
            commit('SET_LOGIN_FLAG', false)
            commit('SET_ACTION', !loginForm.action)
          }
        },
        error => {
          commit('SET_LOGIN_ERROR', new Date().toString() + '服务器异常...')
          commit('SET_LOGIN_FLAG', false)
          commit('SET_ACTION', !loginForm.action)
          console.log(error)
        }
      )
    },
    GetServerList ({ commit }, body) {
      Vue.http.post(body.url, body)
      .then(
        response => {
          if (response.body.result === 0) {
            commit('SET_LIST_ERROR_MSG', '')
            commit('SET_LIST', response.body.data)
          } else {
            commit('SET_LIST_ERROR_MSG', new Date().toString() + response.body.errormsg)
          }
        },
        error => {
          commit('SET_LIST_ERROR_MSG', new Date().toString() + '服务器异常...')
          console.log(error)
        }
      )
    },
    handleOperate ({ commit }, body) {
      Vue.http.post(body.url, body)
      .then(
        response => {
          if (response.body.result === 0) {
            commit('SET_ADD_SERVER_ERROR_MSG', '')
            commit('SET_LIST', response.body.data)
          } else {
            commit('SET_ADD_SERVER_ERROR_MSG', new Date().toString() + response.body.errormsg)
          }
        },
        error => {
          commit('SET_ADD_SERVER_ERROR_MSG', new Date().toString() + '服务器异常...')
          console.log(error)
        }
      )
    },
    AddRemoteServer ({ commit }, body) {
      Vue.http.post(body.url, body)
      .then(
        response => {
          if (response.body.result === 0) {
            commit('SET_ADD_SERVER_ERROR_MSG', '')
            commit('SET_LIST', response.body.data)
          } else {
            commit('SET_ADD_SERVER_ERROR_MSG', response.body.errormsg)
          }
        },
        error => {
          commit('SET_ADD_SERVER_ERROR_MSG', new Date().toString() + '服务器异常...')
          console.log(error)
        }
      )
    }
  }
}
export default loginx

