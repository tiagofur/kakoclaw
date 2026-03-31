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

  const audienceContacts = ref([])
  const audienceContactsTotal = ref(0)
  const audienceContactsPage = ref(1)
  const audienceContactsLoading = ref(false)
  const audienceLists = ref([])
  const audienceListsLoading = ref(false)
  const audienceSegments = ref([])
  const audienceSegmentsLoading = ref(false)
  const audienceDeliveries = ref([])
  const audienceDeliveriesTotal = ref(0)
  const audienceDeliveriesLoading = ref(false)
  const emailCampaigns = ref([])
  const emailCampaignsLoading = ref(false)

  const experimentVariants = ref([])
  const experimentVariantsLoading = ref(false)
  const campaignMemory = ref([])
  const campaignMemoryLoading = ref(false)
  const campaignVersions = ref([])
  const campaignVersionsLoading = ref(false)

  const companyProfiles = ref([])
  const companyProfilesLoading = ref(false)

  const automations = ref([])
  const automationsLoading = ref(false)
  const automationRuns = ref([])

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

  async function fetchAudienceContacts(params = {}) {
    audienceContactsLoading.value = true

    try {
      const response = await marketingService.audienceListContacts(params)
      audienceContacts.value = response.data?.contacts || []
      audienceContactsTotal.value = response.data?.total || 0
      audienceContactsPage.value = response.data?.page || 1
      return response.data
    } catch (error) {
      console.error('Failed to load audience contacts:', error)
      audienceContacts.value = []
      audienceContactsTotal.value = 0
      throw error
    } finally {
      audienceContactsLoading.value = false
    }
  }

  async function createAudienceContact(data) {
    const response = await marketingService.audienceCreateContact(data)
    await fetchAudienceContacts({ page: audienceContactsPage.value })
    return response.data
  }

  async function updateAudienceContact(id, data) {
    const response = await marketingService.audienceUpdateContact(id, data)
    await fetchAudienceContacts({ page: audienceContactsPage.value })
    return response.data
  }

  async function deleteAudienceContact(id) {
    await marketingService.audienceDeleteContact(id)
    await fetchAudienceContacts({ page: audienceContactsPage.value })
  }

  async function importAudienceContacts(formData) {
    const response = await marketingService.audienceImportContacts(formData)
    await fetchAudienceContacts({ page: 1 })
    return response.data
  }

  async function exportAudienceContacts(params = {}) {
    const response = await marketingService.audienceExportContacts(params)
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'contacts.csv')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  }

  async function fetchAudienceLists(params = {}) {
    audienceListsLoading.value = true

    try {
      const response = await marketingService.audienceListLists(params)
      audienceLists.value = response.data?.lists || []
      return response.data
    } catch (error) {
      console.error('Failed to load audience lists:', error)
      audienceLists.value = []
      throw error
    } finally {
      audienceListsLoading.value = false
    }
  }

  async function createAudienceList(data) {
    const response = await marketingService.audienceCreateList(data)
    await fetchAudienceLists()
    return response.data
  }

  async function updateAudienceList(id, data) {
    const response = await marketingService.audienceUpdateList(id, data)
    await fetchAudienceLists()
    return response.data
  }

  async function deleteAudienceList(id) {
    await marketingService.audienceDeleteList(id)
    await fetchAudienceLists()
  }

  async function addAudienceListMember(listId, contactId) {
    const response = await marketingService.audienceAddListMember(listId, contactId)
    return response.data
  }

  async function removeAudienceListMember(listId, contactId) {
    await marketingService.audienceRemoveListMember(listId, contactId)
  }

  async function fetchAudienceSegments(params = {}) {
    audienceSegmentsLoading.value = true

    try {
      const response = await marketingService.audienceListSegments(params)
      audienceSegments.value = response.data?.segments || []
      return response.data
    } catch (error) {
      console.error('Failed to load audience segments:', error)
      audienceSegments.value = []
      throw error
    } finally {
      audienceSegmentsLoading.value = false
    }
  }

  async function createAudienceSegment(data) {
    const response = await marketingService.audienceCreateSegment(data)
    await fetchAudienceSegments()
    return response.data
  }

  async function updateAudienceSegment(id, data) {
    const response = await marketingService.audienceUpdateSegment(id, data)
    await fetchAudienceSegments()
    return response.data
  }

  async function deleteAudienceSegment(id) {
    await marketingService.audienceDeleteSegment(id)
    await fetchAudienceSegments()
  }

  async function previewAudienceSegment(id, rules) {
    try {
      const response = await marketingService.audiencePreviewSegment(id, rules)
      return response.data
    } catch (error) {
      console.error('Failed to preview segment:', error)
      return null
    }
  }

  async function fetchAudienceDeliveries(params = {}) {
    audienceDeliveriesLoading.value = true

    try {
      const response = await marketingService.audienceListDeliveries(params)
      audienceDeliveries.value = response.data?.deliveries || []
      audienceDeliveriesTotal.value = response.data?.total || 0
      return response.data
    } catch (error) {
      console.error('Failed to load audience deliveries:', error)
      audienceDeliveries.value = []
      audienceDeliveriesTotal.value = 0
      throw error
    } finally {
      audienceDeliveriesLoading.value = false
    }
  }

  async function sendAudienceEmail(data) {
    const response = await marketingService.audienceSendToList(data)
    await fetchAudienceDeliveries()
    return response.data
  }

  async function fetchEmailCampaigns(params = {}) {
    emailCampaignsLoading.value = true

    try {
      const response = await marketingService.audienceListCampaigns(params)
      emailCampaigns.value = response.data?.campaigns || []
      return response.data
    } catch (error) {
      console.error('Failed to load email campaigns:', error)
      emailCampaigns.value = []
      throw error
    } finally {
      emailCampaignsLoading.value = false
    }
  }

  async function createEmailCampaign(data) {
    const response = await marketingService.audienceCreateCampaign(data)
    await fetchEmailCampaigns()
    return response.data
  }

  async function updateEmailCampaign(id, data) {
    const response = await marketingService.audienceUpdateCampaign(id, data)
    await fetchEmailCampaigns()
    return response.data
  }

  async function deleteEmailCampaign(id) {
    await marketingService.audienceDeleteCampaign(id)
    await fetchEmailCampaigns()
  }

  async function sendEmailCampaign(id) {
    const response = await marketingService.audienceSendCampaign(id)
    await fetchEmailCampaigns()
    return response.data
  }

  async function fetchExperimentVariants(campaignId) {
    experimentVariantsLoading.value = true
    try {
      const response = await marketingService.audienceListVariants(campaignId)
      experimentVariants.value = response.data?.variants || []
      return response.data
    } catch (error) {
      console.error('Failed to load experiment variants:', error)
      experimentVariants.value = []
      throw error
    } finally {
      experimentVariantsLoading.value = false
    }
  }

  async function createExperimentVariant(campaignId, data) {
    const response = await marketingService.audienceCreateVariant(campaignId, data)
    await fetchExperimentVariants(campaignId)
    return response.data
  }

  async function deleteExperimentVariant(campaignId, variantId) {
    await marketingService.audienceDeleteVariant(campaignId, variantId)
    await fetchExperimentVariants(campaignId)
  }

  async function setVariantWinner(campaignId, variantId) {
    const response = await marketingService.audienceSetWinner(campaignId, variantId)
    await fetchExperimentVariants(campaignId)
    return response.data
  }

  async function fetchVariantMetrics(campaignId, variantId) {
    try {
      const response = await marketingService.audienceVariantMetrics(campaignId, variantId)
      return response.data
    } catch (error) {
      console.error('Failed to load variant metrics:', error)
      return null
    }
  }

  async function fetchAutomations(params = {}) {
    automationsLoading.value = true
    try {
      const response = await marketingService.audienceListAutomations(params)
      automations.value = response.data?.automations || []
      return response.data
    } catch (error) {
      console.error('Failed to load automations:', error)
      automations.value = []
      throw error
    } finally {
      automationsLoading.value = false
    }
  }

  async function createAutomation(data) {
    const response = await marketingService.audienceCreateAutomation(data)
    await fetchAutomations()
    return response.data
  }

  async function deleteAutomation(id) {
    await marketingService.audienceDeleteAutomation(id)
    await fetchAutomations()
  }

  async function fetchAutomationRuns(id) {
    try {
      const response = await marketingService.audienceListAutomationRuns(id)
      automationRuns.value = response.data?.runs || []
      return response.data
    } catch (error) {
      console.error('Failed to load automation runs:', error)
      automationRuns.value = []
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

  async function fetchCompanyProfiles() {
    companyProfilesLoading.value = true
    try {
      const response = await marketingService.listCompanyProfiles()
      companyProfiles.value = response.data?.profiles || []
    } catch (error) {
      console.error('Failed to load company profiles:', error)
      companyProfiles.value = []
    } finally {
      companyProfilesLoading.value = false
    }
  }

  async function getCompanyProfile(account) {
    const response = await marketingService.getCompanyProfile(account)
    return response.data
  }

  async function saveCompanyProfile(account, data) {
    const response = await marketingService.upsertCompanyProfile(account, data)
    const idx = companyProfiles.value.findIndex((p) => p.account === account)
    if (idx >= 0) companyProfiles.value[idx] = response.data
    else companyProfiles.value.push(response.data)
    return response.data
  }

  async function researchCompany(account, website = '') {
    const response = await marketingService.researchCompany(account, website ? { website } : {})
    const idx = companyProfiles.value.findIndex((p) => p.account === account)
    if (idx >= 0) companyProfiles.value[idx] = response.data
    else companyProfiles.value.push(response.data)
    return response.data
  }

  async function fetchCampaignMemory(campaignId) {
    campaignMemoryLoading.value = true
    try {
      const response = await marketingService.audienceListCampaignMemory(campaignId)
      campaignMemory.value = response.data?.entries || []
    } catch (error) {
      console.error('Failed to load campaign memory:', error)
      campaignMemory.value = []
    } finally {
      campaignMemoryLoading.value = false
    }
  }

  async function addCampaignMemoryEntry(campaignId, content, role = 'note') {
    const response = await marketingService.audienceAddCampaignMemoryEntry(campaignId, { content, role })
    campaignMemory.value.push(response.data)
    return response.data
  }

  async function deleteCampaignMemoryEntry(campaignId, entryId) {
    await marketingService.audienceDeleteCampaignMemoryEntry(campaignId, entryId)
    campaignMemory.value = campaignMemory.value.filter((e) => e.id !== entryId)
  }

  async function fetchCampaignVersions(campaignId) {
    campaignVersionsLoading.value = true
    try {
      const response = await marketingService.audienceListVersions(campaignId)
      campaignVersions.value = response.data?.versions || []
    } catch (error) {
      console.error('Failed to load campaign versions:', error)
      campaignVersions.value = []
    } finally {
      campaignVersionsLoading.value = false
    }
  }

  async function saveCampaignVersion(campaignId, note = '') {
    const response = await marketingService.audienceSaveVersion(campaignId, note ? { note } : {})
    campaignVersions.value.unshift(response.data)
    return response.data
  }

  async function restoreCampaignVersion(campaignId, versionId) {
    const response = await marketingService.audienceRestoreVersion(campaignId, versionId)
    await fetchEmailCampaigns()
    return response.data
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
    stopPolling,
    audienceContacts,
    audienceContactsTotal,
    audienceContactsPage,
    audienceContactsLoading,
    audienceLists,
    audienceListsLoading,
    audienceSegments,
    audienceSegmentsLoading,
    audienceDeliveries,
    audienceDeliveriesTotal,
    audienceDeliveriesLoading,
    fetchAudienceContacts,
    createAudienceContact,
    updateAudienceContact,
    deleteAudienceContact,
    importAudienceContacts,
    exportAudienceContacts,
    fetchAudienceLists,
    createAudienceList,
    updateAudienceList,
    deleteAudienceList,
    addAudienceListMember,
    removeAudienceListMember,
    fetchAudienceSegments,
    createAudienceSegment,
    updateAudienceSegment,
    deleteAudienceSegment,
    previewAudienceSegment,
    fetchAudienceDeliveries,
    sendAudienceEmail,
    emailCampaigns,
    emailCampaignsLoading,
    fetchEmailCampaigns,
    createEmailCampaign,
    updateEmailCampaign,
    deleteEmailCampaign,
    sendEmailCampaign,
    experimentVariants,
    experimentVariantsLoading,
    fetchExperimentVariants,
    createExperimentVariant,
    deleteExperimentVariant,
    setVariantWinner,
    fetchVariantMetrics,
    automations,
    automationsLoading,
    automationRuns,
    fetchAutomations,
    createAutomation,
    deleteAutomation,
    fetchAutomationRuns,
    campaignMemory,
    campaignMemoryLoading,
    fetchCampaignMemory,
    addCampaignMemoryEntry,
    deleteCampaignMemoryEntry,
    campaignVersions,
    campaignVersionsLoading,
    fetchCampaignVersions,
    saveCampaignVersion,
    restoreCampaignVersion,
    companyProfiles,
    companyProfilesLoading,
    fetchCompanyProfiles,
    getCompanyProfile,
    saveCompanyProfile,
    researchCompany
  }
})
