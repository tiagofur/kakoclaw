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
  }
}
