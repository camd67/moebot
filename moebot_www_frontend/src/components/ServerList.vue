<template>
     <v-list class="pt-0" dense>
        <v-list-tile v-for="server in serverList" :key="server.id">
            <v-list-tile-action>
                <v-avatar size="32px">
                    <img :src="server.icon"/>
                </v-avatar>
            </v-list-tile-action>
            <v-list-tile-content>
                <v-list-tile-title></v-list-tile-title>
            </v-list-tile-content>
        </v-list-tile>
    </v-list>
</template>

<script>
export default {
  data () {
    return {
      serverList: []
    }
  },
  mounted: function () {
    this.$http.get('/api/serverlist', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('jwt')}}).then(
      response => {
        this.serverList = response.data
      },
      response => {}
    )
  }
}
</script>

<style>

</style>
