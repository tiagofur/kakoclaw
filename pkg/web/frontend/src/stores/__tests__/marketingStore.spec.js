import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useMarketingStore } from '../marketingStore'
import marketingService from '../../services/marketingService'

vi.mock('../../services/marketingService', () => ({
  default: {
    listCampaigns: vi.fn(),
    getCampaign: vi.fn(),
    createCampaign: vi.fn(),
    renameCampaign: vi.fn(),
    deleteCampaign: vi.fn(),
    updateCampaignStatus: vi.fn()
  }
}))

describe('marketingStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('fetchCampaigns() populates campaigns and expands accounts', async () => {
    marketingService.listCampaigns.mockResolvedValue({
      data: {
        campaigns: [
          { account: 'acme', campaign: 'launch', status: 'draft' },
          { account: 'globex', campaign: 'retention', status: 'active' }
        ]
      }
    })

    const store = useMarketingStore()
    await store.fetchCampaigns()

    expect(store.campaigns).toHaveLength(2)
    expect(store.expandedAccounts.has('acme')).toBe(true)
    expect(store.expandedAccounts.has('globex')).toBe(true)
  })

  it('fetchCampaigns() preserves manually collapsed accounts', async () => {
    const campaigns = [
      { account: 'acme', campaign: 'launch', status: 'draft' },
      { account: 'globex', campaign: 'retention', status: 'active' }
    ]

    marketingService.listCampaigns.mockResolvedValue({ data: { campaigns } })

    const store = useMarketingStore()
    await store.fetchCampaigns()
    store.toggleAccount('acme')

    await store.fetchCampaigns()

    expect(store.expandedAccounts.has('acme')).toBe(false)
    expect(store.expandedAccounts.has('globex')).toBe(true)
  })

  it('selectCampaign() stores selection and loads campaign detail', async () => {
    const selected = { account: 'acme', campaign: 'launch', status: 'draft' }
    marketingService.getCampaign.mockResolvedValue({
      data: { ...selected, brief: 'Hello world', files: {} }
    })

    const store = useMarketingStore()
    await store.selectCampaign(selected)

    expect(store.selectedCampaign).toEqual(selected)
    expect(marketingService.getCampaign).toHaveBeenCalledWith('acme', 'launch')
    expect(store.campaignDetail?.brief).toBe('Hello world')
  })

  it('toggleAccount() adds and removes account names', () => {
    const store = useMarketingStore()

    store.toggleAccount('acme')
    expect(store.expandedAccounts.has('acme')).toBe(true)

    store.toggleAccount('acme')
    expect(store.expandedAccounts.has('acme')).toBe(false)
  })

  it('deleteCampaign() clears selection when deleting active campaign', async () => {
    marketingService.deleteCampaign.mockResolvedValue({})
    marketingService.listCampaigns.mockResolvedValue({ data: { campaigns: [] } })

    const store = useMarketingStore()
    store.selectedCampaign = { account: 'acme', campaign: 'launch' }
    store.campaignDetail = { account: 'acme', campaign: 'launch', status: 'draft' }

    await store.deleteCampaign('acme', 'launch')

    expect(store.selectedCampaign).toBe(null)
    expect(store.campaignDetail).toBe(null)
  })

  it('updateCampaignStatus() updates local status optimistically and keeps confirmed value', async () => {
    marketingService.updateCampaignStatus.mockResolvedValue({
      data: { status: 'paused' }
    })

    const store = useMarketingStore()
    store.campaigns = [{ account: 'acme', campaign: 'launch', status: 'draft' }]
    store.selectedCampaign = { account: 'acme', campaign: 'launch', status: 'draft' }
    store.campaignDetail = { account: 'acme', campaign: 'launch', status: 'draft' }

    await store.updateCampaignStatus('acme', 'launch', 'paused')

    expect(store.campaigns[0].status).toBe('paused')
    expect(store.selectedCampaign.status).toBe('paused')
    expect(store.campaignDetail.status).toBe('paused')
  })

  it('updateCampaignStatus() reverts local state when request fails', async () => {
    marketingService.updateCampaignStatus.mockRejectedValue(new Error('boom'))

    const store = useMarketingStore()
    store.campaigns = [{ account: 'acme', campaign: 'launch', status: 'draft' }]
    store.selectedCampaign = { account: 'acme', campaign: 'launch', status: 'draft' }
    store.campaignDetail = { account: 'acme', campaign: 'launch', status: 'draft' }

    await expect(store.updateCampaignStatus('acme', 'launch', 'paused')).rejects.toThrow('boom')

    expect(store.campaigns[0].status).toBe('draft')
    expect(store.selectedCampaign.status).toBe('draft')
    expect(store.campaignDetail.status).toBe('draft')
  })

  it('startPolling() refreshes visible documents and stopPolling() clears the timer', async () => {
    vi.useFakeTimers()
    marketingService.listCampaigns.mockResolvedValue({ data: { campaigns: [] } })
    const clearIntervalSpy = vi.spyOn(globalThis, 'clearInterval')
    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      value: 'visible'
    })

    const store = useMarketingStore()
    store.startPolling()

    await vi.advanceTimersByTimeAsync(10000)
    expect(marketingService.listCampaigns).toHaveBeenCalledTimes(1)

    store.stopPolling()
    expect(clearIntervalSpy).toHaveBeenCalled()

    vi.useRealTimers()
  })
})
