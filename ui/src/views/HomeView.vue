<script setup>
import { ref } from 'vue';
import VideoServices from '@/services/video';

const video = ref();
// onMounted(() => {
//   var options = {
//     width: 1024,
//     height: 768,
//     plugins: {
//       videoJsResolutionSwitcher: {
//         default: 480,
//         dynamicLabel: true
//       }
//     }
//   }

//   const player = window.videojs('content_video', options);

//   player.on(['loadstart', 'play', 'playing', 'firstplay', 'pause', 'ended', 'adplay', 'adplaying', 'adfirstplay', 'adpause', 'adended', 'contentplay', 'contentplaying', 'contentfirstplay', 'contentpause', 'contentended', 'contentupdate', 'loadeddata', 'loadedmetadata'], function (e) {
//     console.warn('VIDEOJS player event: ', e.type);
//   });

// });

const handleSearchFile = async () => {
  const formData = new FormData();
  formData.append('video', video.value.files[0]);
  formData.append('title', 'title_video');
  const res = await VideoServices.postVideo(formData);
  console.log(res);
}

</script>

<template>
  <!-- <video id="content_video" class="video-js vjs-default-skin" controls preload="none" autoplay loop muted width="640"
    height="360">
    <source src="http://localhost:9000/media/34164ac7-7362-41a1-b8a3-ac26830310bb/stream/" type="application/x-mpegURL">
  </video> -->
  <input type="file" id="logo" name="input-logo" @change="handleSearchFile" ref="video">
</template>

<style scoped>
.player {
  position: absolute !important;
  width: 100%;
  height: 60%;
}

.vjs-custom-skin {
  height: 60% !important;
}

.vjs-custom-skin /deep/ .video-js {
  height: 60%;
}
</style>

