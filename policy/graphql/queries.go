package graphql

const (
	// language=graphql
	vulnerabilitiesByPackageQuery = `
	query ($context: Context!, $packageUrls: [String!]!) {
		vulnerabilitiesByPackage(context: $context, packageUrls: $packageUrls) {
			purl
			vulnerabilities {
			cvss {
				severity
				score
			}
			fixedBy
			publishedAt
			source
			sourceId
			updatedAt
			url
			vulnerableRange
			}
		}
	}`

	// language=graphql
	imagePackagesByDigestQuery = `
	query ($context: Context!, $digest: String!) {
		imagePackagesByDigest(context: $context, digest: $digest) {
		  digest
		  imagePackages {
			packages {
			  locations {
				diffId
				path
			  }
			  package {
				licenses
				name
				namespace
				version
				purl
				type
				vulnerabilities {
					cvss {
						severity
						score
					}
					fixedBy
					publishedAt
					source
					sourceId
					updatedAt
					url
					vulnerableRange
				}
			  }
			}
		  }
		  imageHistories {
			emptyLayer
			ordinal
		  }
		  imageLayers {
			layers {
			  diffId
			  ordinal
			}
		  }
		}
	  }
`

	// language=graphql
	baseImagesByDigest = `
	query ($context: Context!, $digest: String!, $platform: ImagePlatform!) {
		imageDetailsByDigest(
			context: $context
			digest: $digest
			platform: $platform
		) {
			digest
			baseImage {
				digest
				repository {
					hostName
					repoName
				}
				tags {
					name
					current
				}
			}
			baseImageTag {
				name
				current
			}
		}
	}
`
)
