package jynx

const (
	ImagePackagesByDigestQueryName = "image-packages-by-digest"

	ImagePackagesByDigestQuery = `
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
					epss {
						percentile
						score
					}
					fixedBy
					publishedAt
					source
					sourceId
					updatedAt
					url
					vulnerableRange
					cisaExploited 
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

	VulnerabilitiesByPackageQueryName = "vulnerabilities-by-package"

	// language=graphql
	VulnerabilitiesByPackageQuery = `
	query ($context: Context!, $purls: [String!]!) {
		vulnerabilitiesByPackage(context: $context, packageUrls: $purls) {
			purl
			vulnerabilities {
			cvss {
				severity
				score
			}
			epss {
				percentile
				score
			}
			fixedBy
			publishedAt
			source
			sourceId
			updatedAt
			url
			vulnerableRange
			cisaExploited 
			}
		}
	}`
)
