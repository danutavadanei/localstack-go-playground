<template>
  <div class="px-4 sm:px-6 lg:px-8">
    <div class="sm:flex sm:items-center">
      <div class="sm:flex-auto">
        <h1 class="text-xl font-semibold text-gray-900">Files from {{ bucket }}</h1>
        <p class="mt-2 text-sm text-gray-700">A list of all the files from this bucket</p>
      </div>
    </div>
    <div class="mt-8 flex flex-col">
      <div class="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
        <div class="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
          <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table class="min-w-full divide-y divide-gray-300">
              <thead class="bg-gray-50">
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-6">Name</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Last Modified</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Size</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900">Storage Class</th>
                  <th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
                    <span class="sr-only">Download</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 bg-white">
                <tr v-for="file in files" :key="file.Key">
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-6">{{ file.Key }}</td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ file.LastModified }}</td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ file.Size }} Bytes</td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">{{ file.StorageClass }}</td>
                  <td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                    <a :href="`http://localhost:8080/s3/buckets/${bucket}/${file.Key}`" target="_blank" class="text-indigo-600 hover:text-indigo-900"
                      >Download<span class="sr-only">, {{ file.Key }}</span></a
                    >
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, inject, onMounted } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const axios = inject('axios')
const files = reactive([])
const bucket = route.params.bucket

onMounted(async () => {
  await axios.get('http://localhost:8080/s3/buckets/' + bucket)
    .then(response => files.push(...response.data.Contents))
})

</script>
