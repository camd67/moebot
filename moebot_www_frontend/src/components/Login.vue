<template>
    <div>
        <PasswordLogin/>
        <DiscordLogin/>
    </div>
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
