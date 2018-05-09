<template>
  <div>
    <v-btn v-on:click="discordLogin">Login with Discord</v-btn>
  </div>
</template>

<script>
export default {
  methods: {
    discordLogin: function (event) {
      window.location.href = '/auth/discord'
    }
  }
}
</script>
