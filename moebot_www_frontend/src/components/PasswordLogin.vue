<template>
  <div>
    <v-form ref="form" @submit.prevent="onSubmit">
      <label class="text-error" v-if="submitError">{{submitError}}</label>
      <v-text-field label="Username" type="text" v-model="user.username"></v-text-field>
      <v-text-field label="Password" type="password" v-model="user.password"></v-text-field>
      <v-btn type="submit">submit</v-btn>
    </v-form>
  </div>
</template>

<script>
export default {
  data () {
    return {
      user: {
        username: '',
        password: ''
      },
      submitError: '',
      submitting: false
    }
  },
  methods: {
    onSubmit: function (event) {
      this.submitting = true
      this.$http.post('/auth/password', this.user).then(
        response => {
          if (response.data && response.data.Jwt) {
            localStorage.setItem('jwt', response.data.Jwt)
            this.$router.push('/')
          }
        },
        response => {
          this.submitting = false
          this.submitError = response.error
        }
      )
    }
  }
}
</script>
