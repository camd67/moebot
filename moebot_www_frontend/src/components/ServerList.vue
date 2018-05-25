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
      serverList: [
        {id: 1, icon: 'https://vuejs.org/images/logo.png'},
        {id: 2, icon: 'https://vuejs.org/images/logo.png'},
        {id: 3, icon: 'https://vuejs.org/images/logo.png'},
        {id: 4, icon: 'https://vuejs.org/images/logo.png'}
      ]
    }
  }
}
</script>

<style>

</style>
