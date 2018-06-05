<template>
  <v-card-actions>
    <v-spacer/>
    <v-btn dark color="grey darken-3" v-on:click="discordLogin">
      <v-avatar size="32px">
        <img src="https://discordapp.com/assets/28174a34e77bb5e5310ced9f95cb480b.png" alt="" />
      </v-avatar>
      Login with Discord
    </v-btn>
  </v-card-actions>
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
