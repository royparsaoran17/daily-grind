/** Signed, direct-to-Cloudinary image upload. Returns the secure URL. */
export function useUpload() {
  async function uploadImage(file: File, kind: 'avatar' | 'post' = 'post'): Promise<string> {
    const sig = await useApi()<{ cloudName: string; apiKey: string; timestamp: number; folder: string; signature: string }>(
      `/uploads/signature?kind=${kind}`,
    )
    const form = new FormData()
    form.append('file', file)
    form.append('api_key', String(sig.apiKey))
    form.append('timestamp', String(sig.timestamp))
    form.append('folder', sig.folder)
    form.append('signature', sig.signature)
    const up = await $fetch<{ secure_url: string }>(
      `https://api.cloudinary.com/v1_1/${sig.cloudName}/image/upload`,
      { method: 'POST', body: form },
    )
    return up.secure_url
  }
  return { uploadImage }
}
