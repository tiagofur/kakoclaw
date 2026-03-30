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
    updateCampaignStatus: vi.fn(),
    fetchTemplates: vi.fn(),
    createTemplate: vi.fn(),
    updateTemplate: vi.fn(),
    deleteTemplate: vi.fn(),
    fetchMedia: vi.fn(),
    uploadMedia: vi.fn(),
    updateMedia: vi.fn(),
    deleteMedia: vi.fn(),
    copyMedia: vi.fn(),
    fetchAnalyticsSummary: vi.fn()
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

  it('fetchTemplates() stores account templates', async () => {
    marketingService.fetchTemplates.mockResolvedValue({
      data: {
        templates: [
          { slug: 'launch-template', name: 'Launch', body: 'Hello {{name}}', variables: ['name'] }
        ]
      }
    })

    const store = useMarketingStore()
    const result = await store.fetchTemplates('acme')

    expect(marketingService.fetchTemplates).toHaveBeenCalledWith('acme')
    expect(result).toHaveLength(1)
    expect(store.templates[0].slug).toBe('launch-template')
  })

  it('createTemplate() refreshes templates after create', async () => {
    marketingService.createTemplate.mockResolvedValue({
      data: { slug: 'launch-template', name: 'Launch', body: 'Hello {{name}}' }
    })
    marketingService.fetchTemplates.mockResolvedValue({
      data: {
        templates: [
          { slug: 'launch-template', name: 'Launch', body: 'Hello {{name}}', variables: ['name'] }
        ]
      }
    })

    const store = useMarketingStore()
    const created = await store.createTemplate('acme', { slug: 'launch-template', name: 'Launch', body: 'Hello {{name}}' })

    expect(marketingService.createTemplate).toHaveBeenCalledWith('acme', { slug: 'launch-template', name: 'Launch', body: 'Hello {{name}}' })
    expect(marketingService.fetchTemplates).toHaveBeenCalledWith('acme')
    expect(created.slug).toBe('launch-template')
  })

  it('fetchMedia() stores account media items', async () => {
    marketingService.fetchMedia.mockResolvedValue({
      data: {
        media: [
          { filename: 'banner.png', path: 'marketing/acme/_media/banner.png', tags: ['launch'] }
        ]
      }
    })

    const store = useMarketingStore()
    const result = await store.fetchMedia('acme')

    expect(marketingService.fetchMedia).toHaveBeenCalledWith('acme')
    expect(result[0].filename).toBe('banner.png')
    expect(store.media[0].tags).toEqual(['launch'])
  })

  it('copyMediaToCampaign() refreshes media after copy', async () => {
    marketingService.copyMedia.mockResolvedValue({
      data: { filename: 'banner.png', campaigns_used: ['launch'] }
    })
    marketingService.fetchMedia.mockResolvedValue({
      data: {
        media: [
          { filename: 'banner.png', campaigns_used: ['launch'] }
        ]
      }
    })

    const store = useMarketingStore()
    const copied = await store.copyMediaToCampaign('acme', 'banner.png', 'launch')

    expect(marketingService.copyMedia).toHaveBeenCalledWith('acme', 'banner.png', 'launch')
    expect(marketingService.fetchMedia).toHaveBeenCalledWith('acme')
    expect(copied.filename).toBe('banner.png')
  })

  it('uploadMedia() refreshes media after upload', async () => {
    marketingService.uploadMedia.mockResolvedValue({
      data: { filename: 'banner.png' }
    })
    marketingService.fetchMedia.mockResolvedValue({
      data: {
        media: [
          { filename: 'banner.png', path: 'marketing/acme/_media/banner.png' }
        ]
      }
    })

    const store = useMarketingStore()
    const formData = new FormData()
    const uploaded = await store.uploadMedia('acme', formData)

    expect(marketingService.uploadMedia).toHaveBeenCalledWith('acme', formData)
    expect(marketingService.fetchMedia).toHaveBeenCalledWith('acme')
    expect(uploaded.filename).toBe('banner.png')
  })

  it('updateMedia() refreshes media after patch', async () => {
    marketingService.updateMedia.mockResolvedValue({
      data: { filename: 'banner.png', tags: ['hero'] }
    })
    marketingService.fetchMedia.mockResolvedValue({
      data: {
        media: [
          { filename: 'banner.png', tags: ['hero'] }
        ]
      }
    })

    const store = useMarketingStore()
    const updated = await store.updateMedia('acme', 'banner.png', { tags: ['hero'], description: 'Hero asset' })

    expect(marketingService.updateMedia).toHaveBeenCalledWith('acme', 'banner.png', { tags: ['hero'], description: 'Hero asset' })
    expect(marketingService.fetchMedia).toHaveBeenCalledWith('acme')
    expect(updated.tags).toEqual(['hero'])
  })

  it('deleteMedia() refreshes media after delete', async () => {
    marketingService.deleteMedia.mockResolvedValue({})
    marketingService.fetchMedia.mockResolvedValue({ data: { media: [] } })

    const store = useMarketingStore()
    await store.deleteMedia('acme', 'banner.png')

    expect(marketingService.deleteMedia).toHaveBeenCalledWith('acme', 'banner.png')
    expect(marketingService.fetchMedia).toHaveBeenCalledWith('acme')
    expect(store.media).toEqual([])
  })

  it('fetchAnalyticsSummary() stores aggregated analytics data', async () => {
    marketingService.fetchAnalyticsSummary.mockResolvedValue({
      data: {
        total_posts: 3,
        total_impressions: 350,
        total_engagements: 60,
        engagement_rate: 60 / 350,
        by_platform: { twitter: { impressions: 150 }, linkedin: { impressions: 200 } },
        trend_dates: ['2026-04-01', '2026-04-02'],
        trend_impressions: [150, 200]
      }
    })

    const store = useMarketingStore()
    const summary = await store.fetchAnalyticsSummary('acme', 'launch')

    expect(marketingService.fetchAnalyticsSummary).toHaveBeenCalledWith('acme', 'launch')
    expect(summary.total_posts).toBe(3)
    expect(store.analyticsSummary.total_impressions).toBe(350)
  })
})
