import { defineStore } from 'pinia'
import { ref } from 'vue'
import marketingService from '../services/marketingService'

export const useMarketingStore = defineStore('marketing', () => {
  const campaigns = ref([])
  const selectedCampaign = ref(null)
  const campaignDetail = ref(null)
  const templates = ref([])
  const media = ref([])
  const analyticsSummary = ref(null)
  const loadingList = ref(false)
  const loadingDetail = ref(false)
  const expandedAccounts = ref(new Set())

  let pollInterval = null

  function matchesCampaign(candidate, account, campaign) {
    return candidate?.account === account && candidate?.campaign === campaign
  }

  function syncSelectedCampaign(nextCampaigns) {
    if (!selectedCampaign.value) return

    const nextSelected = nextCampaigns.find((campaign) => matchesCampaign(campaign, selectedCampaign.value.account, selectedCampaign.value.campaign))

    if (nextSelected) {
      selectedCampaign.value = nextSelected
      return
    }

    selectedCampaign.value = null
    campaignDetail.value = null
  }

  function expandKnownAccounts(nextCampaigns, previousCampaigns = []) {
    const knownAccounts = new Set(previousCampaigns.map((campaign) => campaign.account))
    const nextExpanded = new Set(expandedAccounts.value)
    nextCampaigns.forEach((campaign) => {
      if (!knownAccounts.has(campaign.account)) {
        nextExpanded.add(campaign.account)
      }
    })
    expandedAccounts.value = nextExpanded
  }

  function setLocalCampaignStatus(account, campaign, status) {
    campaigns.value = campaigns.value.map((item) => {
      if (!matchesCampaign(item, account, campaign)) return item
      return { ...item, status }
    })

    if (matchesCampaign(selectedCampaign.value, account, campaign)) {
      selectedCampaign.value = { ...selectedCampaign.value, status }
    }

    if (matchesCampaign(campaignDetail.value, account, campaign)) {
      campaignDetail.value = { ...campaignDetail.value, status }
    }
  }

  async function fetchCampaigns() {
    loadingList.value = true

    try {
      const previousCampaigns = campaigns.value
      const response = await marketingService.listCampaigns()
      const nextCampaigns = response.data?.campaigns || []

      campaigns.value = nextCampaigns
      expandKnownAccounts(nextCampaigns, previousCampaigns)
      syncSelectedCampaign(nextCampaigns)

      return nextCampaigns
    } catch (error) {
      console.error('Failed to load campaigns:', error)
      campaigns.value = []
      syncSelectedCampaign([])
      return []
    } finally {
      loadingList.value = false
    }
  }

  async function fetchCampaignDetail(campaign = selectedCampaign.value) {
    if (!campaign) {
      campaignDetail.value = null
      analyticsSummary.value = null
      return null
    }

    loadingDetail.value = true
    campaignDetail.value = null

    try {
      const response = await marketingService.getCampaign(campaign.account, campaign.campaign)
      campaignDetail.value = response.data
      return response.data
    } catch (error) {
      console.error('Failed to load campaign detail:', error)
      analyticsSummary.value = null
      return null
    } finally {
      loadingDetail.value = false
    }
  }

  async function selectCampaign(campaign) {
    selectedCampaign.value = campaign
    return fetchCampaignDetail(campaign)
  }

  function toggleAccount(account) {
    const nextExpanded = new Set(expandedAccounts.value)

    if (nextExpanded.has(account)) {
      nextExpanded.delete(account)
    } else {
      nextExpanded.add(account)
    }

    expandedAccounts.value = nextExpanded
  }

  async function createCampaign(payload) {
    const response = await marketingService.createCampaign(payload)
    await fetchCampaigns()
    return response.data
  }

  async function renameCampaign(account, campaign, newName) {
    const response = await marketingService.renameCampaign(account, campaign, newName)
    const renamedCampaign = response.data

    await fetchCampaigns()

    if (matchesCampaign(selectedCampaign.value, account, campaign)) {
      selectedCampaign.value = renamedCampaign
      await fetchCampaignDetail(renamedCampaign)
    }

    return renamedCampaign
  }

  async function deleteCampaign(account, campaign) {
    const wasSelected = matchesCampaign(selectedCampaign.value, account, campaign)

    await marketingService.deleteCampaign(account, campaign)
    await fetchCampaigns()

    if (wasSelected) {
      selectedCampaign.value = null
      campaignDetail.value = null
      analyticsSummary.value = null
    }
  }

  async function updateCampaignStatus(account, campaign, status) {
    const previousStatus = campaignDetail.value?.status || selectedCampaign.value?.status || campaigns.value.find((item) => matchesCampaign(item, account, campaign))?.status || 'draft'

    setLocalCampaignStatus(account, campaign, status)

    try {
      const response = await marketingService.updateCampaignStatus(account, campaign, status)
      const confirmedStatus = response.data?.status || status
      setLocalCampaignStatus(account, campaign, confirmedStatus)
      return response.data
    } catch (error) {
      setLocalCampaignStatus(account, campaign, previousStatus)
      throw error
    }
  }

  async function fetchTemplates(account) {
    if (!account) {
      templates.value = []
      return []
    }

    try {
      const response = await marketingService.fetchTemplates(account)
      templates.value = response.data?.templates || []
      return templates.value
    } catch (error) {
      console.error('Failed to load templates:', error)
      templates.value = []
      throw error
    }
  }

  async function createTemplate(account, payload) {
    const response = await marketingService.createTemplate(account, payload)
    await fetchTemplates(account)
    return response.data
  }

  async function updateTemplate(account, slug, payload) {
    const response = await marketingService.updateTemplate(account, slug, payload)
    await fetchTemplates(account)
    return response.data
  }

  async function deleteTemplate(account, slug) {
    await marketingService.deleteTemplate(account, slug)
    await fetchTemplates(account)
  }

  async function fetchMedia(account) {
    if (!account) {
      media.value = []
      return []
    }

    try {
      const response = await marketingService.fetchMedia(account)
      media.value = response.data?.media || []
      return media.value
    } catch (error) {
      console.error('Failed to load media:', error)
      media.value = []
      throw error
    }
  }

  async function uploadMedia(account, formData) {
    const response = await marketingService.uploadMedia(account, formData)
    await fetchMedia(account)
    return response.data
  }

  async function updateMedia(account, filename, payload) {
    const response = await marketingService.updateMedia(account, filename, payload)
    await fetchMedia(account)
    return response.data
  }

  async function deleteMedia(account, filename) {
    await marketingService.deleteMedia(account, filename)
    await fetchMedia(account)
  }

  async function copyMedia(account, filename, campaign) {
    const response = await marketingService.copyMedia(account, filename, campaign)
    await fetchMedia(account)
    return response.data
  }

  async function copyMediaToCampaign(account, filename, campaign) {
    return copyMedia(account, filename, campaign)
  }

  async function fetchAnalyticsSummary(account, campaign) {
    if (!account || !campaign) {
      analyticsSummary.value = null
      return null
    }

    try {
      const response = await marketingService.fetchAnalyticsSummary(account, campaign)
      analyticsSummary.value = response.data || null
      return analyticsSummary.value
    } catch (error) {
      console.error('Failed to load analytics summary:', error)
      analyticsSummary.value = null
      throw error
    }
  }

  function startPolling() {
    stopPolling()

    pollInterval = setInterval(() => {
      if (typeof document === 'undefined' || document.visibilityState === 'visible') {
        fetchCampaigns()
      }
    }, 10000)
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  return {
    campaigns,
    selectedCampaign,
    campaignDetail,
    templates,
    media,
    analyticsSummary,
    loadingList,
    loadingDetail,
    expandedAccounts,
    fetchCampaigns,
    fetchCampaignDetail,
    selectCampaign,
    toggleAccount,
    createCampaign,
    renameCampaign,
    deleteCampaign,
    updateCampaignStatus,
    fetchTemplates,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    fetchMedia,
    uploadMedia,
    updateMedia,
    deleteMedia,
    copyMedia,
    copyMediaToCampaign,
    fetchAnalyticsSummary,
    startPolling,
    stopPolling
  }
})
