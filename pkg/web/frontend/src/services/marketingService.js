import client from './api'

const encodeSegment = (value = '') => encodeURIComponent(String(value))

export default {
  listCampaigns() {
    return client.get('/marketing/campaigns')
  },
  getCampaign(account, campaign) {
    return client.get(`/marketing/campaigns/${encodeSegment(account)}/${encodeSegment(campaign)}`)
  },
  createCampaign(payload) {
    return client.post('/marketing/campaigns', payload)
  },
  renameCampaign(account, campaign, newName) {
    return client.patch(`/marketing/campaigns/${encodeSegment(account)}/${encodeSegment(campaign)}`, { new_name: newName })
  },
  deleteCampaign(account, campaign) {
    return client.delete(`/marketing/campaigns/${encodeSegment(account)}/${encodeSegment(campaign)}`)
  },
  updateCampaignStatus(account, campaign, status) {
    return client.patch(`/marketing/campaigns/${encodeSegment(account)}/${encodeSegment(campaign)}/status`, { status })
  },
  fetchTemplates(account) {
    return client.get(`/marketing/templates/${encodeSegment(account)}`)
  },
  createTemplate(account, payload) {
    return client.post(`/marketing/templates/${encodeSegment(account)}`, payload)
  },
  updateTemplate(account, slug, payload) {
    return client.put(`/marketing/templates/${encodeSegment(account)}/${encodeSegment(slug)}`, payload)
  },
  deleteTemplate(account, slug) {
    return client.delete(`/marketing/templates/${encodeSegment(account)}/${encodeSegment(slug)}`)
  },
  fetchMedia(account) {
    return client.get(`/marketing/media/${encodeSegment(account)}`)
  },
  uploadMedia(account, formData) {
    return client.post(`/marketing/media/${encodeSegment(account)}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },
  updateMedia(account, filename, data) {
    return client.patch(`/marketing/media/${encodeSegment(account)}/${encodeSegment(filename)}`, data)
  },
  deleteMedia(account, filename) {
    return client.delete(`/marketing/media/${encodeSegment(account)}/${encodeSegment(filename)}`)
  },
  copyMedia(account, filename, campaign) {
    return client.post(`/marketing/media/${encodeSegment(account)}/copy`, {
      file: filename,
      campaign
    })
  },
  fetchAnalyticsSummary(account, campaign) {
    return client.get(`/marketing/campaigns/${encodeSegment(account)}/${encodeSegment(campaign)}/analytics/summary`)
  },

  audienceListContacts(params) {
    return client.get('/marketing/audience/contacts', { params })
  },
  audienceGetContact(id) {
    return client.get(`/marketing/audience/contacts/${encodeSegment(id)}`)
  },
  audienceCreateContact(data) {
    return client.post('/marketing/audience/contacts', data)
  },
  audienceUpdateContact(id, data) {
    return client.put(`/marketing/audience/contacts/${encodeSegment(id)}`, data)
  },
  audienceDeleteContact(id) {
    return client.delete(`/marketing/audience/contacts/${encodeSegment(id)}`)
  },
  audienceImportContacts(formData) {
    return client.post('/marketing/audience/contacts/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  audienceExportContacts(params) {
    return client.get('/marketing/audience/contacts/export', { params, responseType: 'blob' })
  },
  audienceListLists(params) {
    return client.get('/marketing/audience/lists', { params })
  },
  audienceCreateList(data) {
    return client.post('/marketing/audience/lists', data)
  },
  audienceUpdateList(id, data) {
    return client.put(`/marketing/audience/lists/${encodeSegment(id)}`, data)
  },
  audienceDeleteList(id) {
    return client.delete(`/marketing/audience/lists/${encodeSegment(id)}`)
  },
  audienceAddListMember(listId, contactId) {
    return client.post(`/marketing/audience/lists/${encodeSegment(listId)}/members`, { contact_id: contactId })
  },
  audienceRemoveListMember(listId, contactId) {
    return client.delete(`/marketing/audience/lists/${encodeSegment(listId)}/members/${encodeSegment(contactId)}`)
  },
  audienceListSegments(params) {
    return client.get('/marketing/audience/segments', { params })
  },
  audienceCreateSegment(data) {
    return client.post('/marketing/audience/segments', data)
  },
  audienceUpdateSegment(id, data) {
    return client.put(`/marketing/audience/segments/${encodeSegment(id)}`, data)
  },
  audienceDeleteSegment(id) {
    return client.delete(`/marketing/audience/segments/${encodeSegment(id)}`)
  },
  audiencePreviewSegment(id, rules) {
    return client.post(`/marketing/audience/segments/${encodeSegment(id)}/preview`, { rules })
  },
  audienceListDeliveries(params) {
    return client.get('/marketing/audience/deliveries', { params })
  },
  audienceSendToList(data) {
    return client.post('/marketing/audience/send', data)
  },

  audienceListCampaigns(params) {
    return client.get('/marketing/audience/email-campaigns', { params })
  },
  audienceCreateCampaign(data) {
    return client.post('/marketing/audience/email-campaigns', data)
  },
  audienceGetCampaign(id) {
    return client.get(`/marketing/audience/email-campaigns/${encodeSegment(id)}`)
  },
  audienceUpdateCampaign(id, data) {
    return client.put(`/marketing/audience/email-campaigns/${encodeSegment(id)}`, data)
  },
  audienceDeleteCampaign(id) {
    return client.delete(`/marketing/audience/email-campaigns/${encodeSegment(id)}`)
  },
  audienceSendCampaign(id) {
    return client.post(`/marketing/audience/email-campaigns/${encodeSegment(id)}/send`)
  },

  audienceListVariants(campaignId) {
    return client.get(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants`)
  },
  audienceCreateVariant(campaignId, data) {
    return client.post(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants`, data)
  },
  audienceUpdateVariant(campaignId, variantId, data) {
    return client.put(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants/${encodeSegment(variantId)}`, data)
  },
  audienceDeleteVariant(campaignId, variantId) {
    return client.delete(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants/${encodeSegment(variantId)}`)
  },
  audienceVariantMetrics(campaignId, variantId) {
    return client.get(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants/${encodeSegment(variantId)}/metrics`)
  },
  audienceSetWinner(campaignId, variantId) {
    return client.post(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/variants/${encodeSegment(variantId)}/winner`)
  },

  listCompanyProfiles() {
    return client.get('/marketing/profiles')
  },
  getCompanyProfile(account) {
    return client.get(`/marketing/profiles/${encodeSegment(account)}`)
  },
  upsertCompanyProfile(account, data) {
    return client.put(`/marketing/profiles/${encodeSegment(account)}`, data)
  },
  researchCompany(account, data = {}) {
    return client.post(`/marketing/profiles/${encodeSegment(account)}/research`, data, { timeout: 120000 })
  },

  audienceCampaignProgress(campaignId) {
    return client.get(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/progress`)
  },

  audienceListCampaignMemory(campaignId) {
    return client.get(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/memory`)
  },
  audienceAddCampaignMemoryEntry(campaignId, data) {
    return client.post(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/memory`, data)
  },
  audienceDeleteCampaignMemoryEntry(campaignId, entryId) {
    return client.delete(`/marketing/audience/email-campaigns/${encodeSegment(campaignId)}/memory/${encodeSegment(entryId)}`)
  },

  audienceListAutomations(params) {
    return client.get('/marketing/audience/automations', { params })
  },
  audienceCreateAutomation(data) {
    return client.post('/marketing/audience/automations', data)
  },
  audienceGetAutomation(id) {
    return client.get(`/marketing/audience/automations/${encodeSegment(id)}`)
  },
  audienceUpdateAutomation(id, data) {
    return client.put(`/marketing/audience/automations/${encodeSegment(id)}`, data)
  },
  audienceDeleteAutomation(id) {
    return client.delete(`/marketing/audience/automations/${encodeSegment(id)}`)
  },
  audienceListAutomationRuns(id) {
    return client.get(`/marketing/audience/automations/${encodeSegment(id)}/runs`)
  }
}
