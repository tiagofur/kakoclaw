import client from './api'

export default {
  listCampaigns() {
    return client.get('/marketing/campaigns')
  },
  getCampaign(account, campaign) {
    return client.get(`/marketing/campaigns/${account}/${campaign}`)
  },
  createCampaign(payload) {
    return client.post('/marketing/campaigns', payload)
  },
  renameCampaign(account, campaign, newName) {
    return client.patch(`/marketing/campaigns/${account}/${campaign}`, { new_name: newName })
  },
  deleteCampaign(account, campaign) {
    return client.delete(`/marketing/campaigns/${account}/${campaign}`)
  },
  updateCampaignStatus(account, campaign, status) {
    return client.patch(`/marketing/campaigns/${account}/${campaign}/status`, { status })
  }
}
