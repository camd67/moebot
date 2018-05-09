<template>
  <div>
    <h1>Welcome to MoeBot!</h1>
    <v-btn color="info" v-on:click="logout">Logout</v-btn>
  </div>
</template>

<script>
export default {
  methods: {
    logout: function (event) {
      localStorage.removeItem('jwt')
      this.$router.go(this.$router.currentRoute)
    }
  }
}
</script>

<style>

</style>
