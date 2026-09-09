import { onBeforeUnmount, ref } from 'vue'

const preferredMimeTypes = [
  'audio/webm;codecs=opus',
  'audio/webm',
  'audio/mp4',
]

export function useAudioRecorder() {
  const isRecording = ref(false)
  const audioBlob = ref<Blob | null>(null)

  let mediaRecorder: MediaRecorder | null = null
  let mediaStream: MediaStream | null = null
  let audioChunks: Blob[] = []
  let isCancelled = false

  function releaseMicrophone() {
    mediaStream?.getTracks().forEach((track) => track.stop())
    mediaStream = null
  }

  function getSupportedMimeType(): string {
    return (
      preferredMimeTypes.find((type) =>
        MediaRecorder.isTypeSupported(type),
      ) ?? ''
    )
  }

  async function startRecording() {
    if (
      !navigator.mediaDevices?.getUserMedia ||
      typeof MediaRecorder === 'undefined'
    ) {
      throw new Error('当前浏览器不支持录音')
    }

    audioBlob.value = null
    audioChunks = []
    isCancelled = false

    mediaStream = await navigator.mediaDevices.getUserMedia({
      audio: true,
    })

    const mimeType = getSupportedMimeType()
    mediaRecorder = mimeType
      ? new MediaRecorder(mediaStream, { mimeType })
      : new MediaRecorder(mediaStream)

    mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) {
        audioChunks.push(event.data)
      }
    }

    mediaRecorder.onstop = () => {
      if (!isCancelled && audioChunks.length > 0) {
        audioBlob.value = new Blob(audioChunks, {
          type: mediaRecorder?.mimeType || mimeType || 'audio/webm',
        })
      }

      audioChunks = []
      mediaRecorder = null
      releaseMicrophone()
    }

    mediaRecorder.start(100)
    isRecording.value = true
  }

  function stopRecording() {
    if (mediaRecorder?.state === 'recording') {
      mediaRecorder.stop()
    }

    isRecording.value = false
  }

  function cancelRecording() {
    isCancelled = true
    audioChunks = []
    audioBlob.value = null

    if (mediaRecorder?.state === 'recording') {
      mediaRecorder.stop()
    } else {
      releaseMicrophone()
    }

    isRecording.value = false
  }

  function clearAudio() {
    audioBlob.value = null
  }

  onBeforeUnmount(() => {
    cancelRecording()
  })

  return {
    isRecording,
    audioBlob,
    startRecording,
    stopRecording,
    cancelRecording,
    clearAudio,
  }
}