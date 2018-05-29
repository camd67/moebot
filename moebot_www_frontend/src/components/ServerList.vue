<template>
     <v-list class="pt-0" dense>
        <v-divider/>
        <v-list-tile v-bind:class="{ loading: isLoading }" v-for="server in serverList" :key="server.id" @click="test">
            <v-list-tile-action>
              <v-tooltip right>
                <v-avatar slot="activator" size="32px">
                    <img :src="server.icon"/>
                </v-avatar>
                <span>{{server.name}}</span>
              </v-tooltip>
            </v-list-tile-action>
        </v-list-tile>
    </v-list>
</template>

<script>
export default {
  data () {
    return {
      isLoading: false,
      serverList: [
        {id: 1, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 2, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 3, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 4, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 5, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 6, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'},
        {id: 7, icon: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAQAAADZc7J/AAAAIUlEQVR42mPM+c9AEWAcNWDUgFEDRg0YNWDUgFEDhpsBAGTCLYFg4onKAAAAAElFTkSuQmCC'}
      ]
    }
  },
  mounted: function () {
    this.isLoading = true
    this.$http.get('/api/serverlist', {headers: {'Authorization': 'Bearer ' + localStorage.getItem('jwt')}}).then(
      response => {
        this.isLoading = false
        this.serverList = response.data
      },
      response => {
        this.serverList = [{id: 1, icon: '/static/discordError.png'}]
        this.isLoading = false
      }
    )
  },
  methods: {
    test: function () { }
  }
}
</script>

<style>
.loading {
  animation-name: loadingPulse;
  animation-duration: 1.5s;
  animation-iteration-count: infinite;
  animation-timing-function: linear;
}

@keyframes loadingPulse {
    0% {opacity: 0.5;}
    50% {opacity: 1;}
    100% {opacity: 0.5;}
}
</style>
