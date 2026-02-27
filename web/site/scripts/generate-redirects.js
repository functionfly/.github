#!/usr/bin/env node

/**
 * Generate _redirects file from Sanity redirects
 * Run this during build process to create Netlify/Vercel compatible redirects
 */

import { createClient } from '@sanity/client'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// Initialize Sanity client
const sanityClient = createClient({
  projectId: process.env.SANITY_PROJECT_ID,
  dataset: process.env.SANITY_DATASET || 'production',
  apiVersion: '2024-01-01',
  useCdn: true, // Use CDN for build-time generation
})

async function generateRedirects() {
  console.log('🔄 Generating redirects from Sanity...')

  try {
    // Fetch all enabled redirects, ordered by source path for consistency
    const redirects = await sanityClient.fetch(`
      *[_type == "redirect" && enabled == true] | order(source asc) {
        source,
        destination,
        statusCode,
        matchType
      }
    `)

    // Generate Netlify/Vercel compatible _redirects format
    const redirectLines = []

    // Add header comment
    redirectLines.push('# Redirects generated from Sanity CMS')
    redirectLines.push(`# Generated at: ${new Date().toISOString()}`)
    redirectLines.push('')

    for (const redirect of redirects) {
      let source = redirect.source

      // Handle different match types
      if (redirect.matchType === 'prefix') {
        // Add trailing slash for prefix matches
        source = source.endsWith('/') ? source : `${source}/*`
      } else if (redirect.matchType === 'regex') {
        // For regex matches, we'll use a simple approach
        // More complex regex handling might need custom logic
        console.warn(`⚠️  Regex redirect not fully supported: ${source}`)
        continue
      }

      // Format: source destination statusCode
      redirectLines.push(`${source} ${redirect.destination} ${redirect.statusCode}`)
    }

    // Add final newline
    redirectLines.push('')

    // Write to _redirects file
    const redirectsPath = path.join(__dirname, '..', 'public', '_redirects')
    fs.writeFileSync(redirectsPath, redirectLines.join('\n'))

    console.log(`✅ Generated ${redirects.length} redirects and saved to ${redirectsPath}`)

  } catch (error) {
    console.error('❌ Error generating redirects:', error)
    process.exit(1)
  }
}

generateRedirects()