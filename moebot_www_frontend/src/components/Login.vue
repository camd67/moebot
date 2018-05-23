<template>
    <v-content>
      <v-container fluid fill-height>
        <v-layout align-center justify-center>
          <v-flex  xs12 sm8 md4>
            <v-card elevation-12>
              <v-toolbar dark color="primary">
                <v-toolbar-title>Login to MoeBot</v-toolbar-title>
              </v-toolbar>
              <PasswordLogin/>
              <DiscordLogin/>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
</template>

<script>
import PasswordLogin from './PasswordLogin.vue'
import DiscordLogin from './DiscordLogin.vue'
export default {
  components: {
    PasswordLogin,
    DiscordLogin
  }
}
</script>
