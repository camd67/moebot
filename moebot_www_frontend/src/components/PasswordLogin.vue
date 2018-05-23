<template>
  <v-form ref="form" @submit.prevent="onSubmit" v-model="valid">
      <v-card-text>
        <label class="text-error" v-if="submitError">{{submitError}}</label>
        <v-text-field
          prepend-icon="person"
          label="Username"
          type="text"
          v-model="user.username"
          :rules="usernameRules"
          required></v-text-field>
        <v-text-field
          prepend-icon="lock"
          label="Password"
          type="password"
          v-model="user.password"
          :rules="passwordRules"
          required></v-text-field>
        <transition name="slide-y-transition">
          <div v-if="register">
            <v-text-field prepend-icon="lock"
              label="Confirm Password"
              type="password"
              :rules="confirmPasswordRules"
              required></v-text-field>
            <v-text-field
              prepend-icon="email"
              label="Email"
              type="text"
              v-model="user.email"
              :rules="emailRules"
              required></v-text-field>
            <v-text-field
              prepend-icon="email"
              label="Confirm Email"
              type="text"
              :rules="confirmEmailRules"
              required></v-text-field>
          </div>
        </transition>
      </v-card-text>
    <v-card-actions>
      <v-spacer/>
      <v-btn :disabled="!valid || submitting" color="primary" type="submit">{{register ? 'Register' : 'Login'}}</v-btn>
      <v-btn color="primary" v-on:click="register = !register">{{register ? 'Sign In' : 'Sign Up'}}</v-btn>
    </v-card-actions>
  </v-form>
</template>

<script>
export default {
  data () {
    return {
      valid: false,
      user: {
        username: '',
        password: '',
        email: ''
      },
      submitError: '',
      submitting: false,
      register: false,
      usernameRules: [u => !!u || 'Username is required'],
      passwordRules: [p => !!p || 'Password is required'],
      emailRules: [
        e => !!e || 'Email is required',
        e => /^\w+([.-]?\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(e) || 'E-mail must be valid'],
      confirmPasswordRules: [cp => cp === this.user.password || 'Password must match'],
      confirmEmailRules: [ce => ce === this.user.email || 'E-mail must match']
    }
  },
  methods: {
    onSubmit: function (event) {
      if (!this.valid) return
      if (this.submitting) return
      this.submitting = true
      this.$http.post(this.register ? '/auth/register' : '/auth/password', this.user).then(
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
