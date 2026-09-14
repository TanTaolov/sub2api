import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import AccountTestModal from '../AccountTestModal.vue'

const { getAvailableModels, copyToClipboard } = vi.hoisted(() => ({
  getAvailableModels: vi.fn(),
  copyToClipboard: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getAvailableModels
    }
  }
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const messages: Record<string, string> = {
    'admin.accounts.imagePromptDefault': 'Generate a cute orange cat astronaut sticker on a clean pastel background.'
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string | number>) => {
        if (key === 'admin.accounts.imageReceived' && params?.count) {
          return `received-${params.count}`
        }
        if (key === 'admin.accounts.imagePreviewAlt' && params?.index) {
          return `test-image-${params.index}`
        }
        return messages[key] || key
      }
    })
  }
})

function createStreamResponse(lines: string[]) {
  const encoder = new TextEncoder()
  const chunks = lines.map((line) => encoder.encode(line))
  let index = 0

  return {
    ok: true,
    body: {
      getReader: () => ({
        read: vi.fn().mockImplementation(async () => {
          if (index < chunks.length) {
            return { done: false, value: chunks[index++] }
          }
          return { done: true, value: undefined }
        })
      })
    }
  } as Response
}

function mountModal(account: Record<string, unknown> = {
  id: 42,
  name: 'Gemini Image Test',
  platform: 'gemini',
  type: 'apikey',
  status: 'active'
}) {
  return mount(AccountTestModal, {
    props: {
      show: false,
      account
    } as any,
    global: {
      stubs: {
        BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
        Select: {
          props: ['modelValue', 'options', 'valueKey', 'labelKey', 'disabled'],
          emits: ['update:modelValue'],
          template: `
            <select class="select-stub" :value="modelValue" :disabled="disabled"
              @change="$emit('update:modelValue', $event.target.value)">
              <option v-for="option in options" :key="option[valueKey || 'value']"
                :value="option[valueKey || 'value']">{{ option[labelKey || 'label'] }}</option>
            </select>
          `
        },
        TextArea: {
          props: ['modelValue'],
          emits: ['update:modelValue'],
          template: '<textarea class="textarea-stub" :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)" />'
        },
        Icon: true
      }
    }
  })
}

describe('AccountTestModal', () => {
  beforeEach(() => {
    getAvailableModels.mockResolvedValue([
      { id: 'gemini-2.0-flash', display_name: 'Gemini 2.0 Flash' },
      { id: 'gemini-2.5-flash-image', display_name: 'Gemini 2.5 Flash Image' },
      { id: 'gemini-3.1-flash-image', display_name: 'Gemini 3.1 Flash Image' }
    ])
    copyToClipboard.mockReset()
    Object.defineProperty(globalThis, 'localStorage', {
      value: {
        getItem: vi.fn((key: string) => (key === 'auth_token' ? 'test-token' : null)),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn()
      },
      configurable: true
    })
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gemini-2.5-flash-image"}\n',
        'data: {"type":"image","image_url":"data:image/png;base64,QUJD","mime_type":"image/png"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('gemini 图片模型测试会携带提示词并渲染图片预览', async () => {
    const wrapper = mountModal()
    await wrapper.setProps({ show: true })
    await flushPromises()

    const promptInput = wrapper.find('textarea.textarea-stub')
    expect(promptInput.exists()).toBe(true)
    await promptInput.setValue('draw a tiny orange cat astronaut')

    const buttons = wrapper.findAll('button')
    const startButton = buttons.find((button) => button.text().includes('admin.accounts.startTest'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await flushPromises()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toEqual({
      model_id: 'gemini-3.1-flash-image',
      prompt: 'draw a tiny orange cat astronaut'
    })

    const preview = wrapper.find('img[alt="test-image-1"]')
    expect(preview.exists()).toBe(true)
    expect(preview.attributes('src')).toBe('data:image/png;base64,QUJD')
  })

  it('grok 账号测试默认选择 Grok 模型', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'grok-4.3', display_name: 'Grok 4.3' },
      { id: 'grok-build-0.1', display_name: 'Grok Build 0.1' }
    ])
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"grok-4.3"}\n',
        'data: {"type":"content","text":"ok"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mountModal({
      id: 13,
      name: 'Grok Account',
      platform: 'grok',
      type: 'oauth',
      status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    const buttons = wrapper.findAll('button')
    const startButton = buttons.find((button) => button.text().includes('admin.accounts.startTest'))
    expect(startButton).toBeTruthy()

    await startButton!.trigger('click')
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toEqual({
      model_id: 'grok-4.3',
      prompt: '',
      mode: 'text'
    })
  })

  it('OpenAI Compact 探测会携带 compact 测试模式', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mountModal({
      id: 42,
      name: 'OpenAI OAuth',
      platform: 'openai',
      type: 'oauth',
      status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    ;(wrapper.vm as any).selectedModelId = 'gpt-5.4'
    ;(wrapper.vm as any).testMode = 'compact'
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toMatchObject({
      model_id: 'gpt-5.4',
      prompt: '',
      mode: 'compact'
    })
  })

  it('OpenAI 糖果测试会携带 candy 模式并显示固定题面的摘要和日志', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    global.fetch = vi.fn().mockResolvedValue(
      createStreamResponse([
        'data: {"type":"test_start","model":"gpt-5.4"}\n',
        'data: {"type":"test_complete","success":true}\n'
      ])
    ) as any

    const wrapper = mountModal({
      id: 43,
      name: 'OpenAI OAuth',
      platform: 'openai',
      type: 'oauth',
      status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    await wrapper.findAll('select.select-stub')[1].setValue('candy')
    expect(wrapper.text()).toContain('admin.accounts.openai.testModeCandySummary')
    expect(wrapper.text()).not.toContain('admin.accounts.testPrompt')
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(global.fetch).toHaveBeenCalledTimes(1)
    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toMatchObject({
      model_id: 'gpt-5.4',
      prompt: '',
      mode: 'candy'
    })
    expect(wrapper.text()).toContain('admin.accounts.openai.sendingCandyTestMessage')
    expect(wrapper.text()).not.toContain('admin.accounts.sendingTestMessage')
  })

  it('切换糖果模式会过滤图片模型并替换旧的图片选择，常规模式仍可生图', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'gpt-image-2', display_name: 'GPT Image 2' },
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    const wrapper = mountModal({
      id: 44, name: 'OpenAI API Key', platform: 'openai', type: 'apikey', status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()

    const [modelSelect, modeSelect] = wrapper.findAll('select.select-stub')
    expect((modelSelect.element as HTMLSelectElement).value).toBe('gpt-image-2')
    await modeSelect.setValue('candy')
    expect(modelSelect.findAll('option').map((option) => option.attributes('value'))).toEqual(['gpt-5.4'])
    expect((modelSelect.element as HTMLSelectElement).value).toBe('gpt-5.4')
    expect(wrapper.find('textarea.textarea-stub').exists()).toBe(false)

    await modeSelect.setValue('default')
    expect(modelSelect.findAll('option').map((option) => option.attributes('value'))).toContain('gpt-image-2')
    await modelSelect.setValue('gpt-image-2')
    await wrapper.find('textarea.textarea-stub').setValue('draw a tiny cat')
    await (wrapper.vm as any).startTest()
    await flushPromises()

    const [, request] = (global.fetch as any).mock.calls[0]
    expect(JSON.parse(request.body)).toEqual({
      model_id: 'gpt-image-2', prompt: 'draw a tiny cat', mode: 'default'
    })
    expect(wrapper.text()).toContain('admin.accounts.sendingImageRequest')
  })

  it('糖果模式没有文本模型时清空旧选择并禁止请求', async () => {
    getAvailableModels.mockResolvedValue([
      { id: 'gpt-image-2', display_name: 'GPT Image 2' }
    ])
    const wrapper = mountModal({
      id: 45, name: 'OpenAI Images', platform: 'openai', type: 'apikey', status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()
    await wrapper.findAll('select.select-stub')[1].setValue('candy')

    expect((wrapper.vm as any).selectedModelId).toBe('')
    const startButton = wrapper.findAll('button').find((button) => button.text().includes('admin.accounts.startTest'))
    expect(startButton!.attributes('disabled')).toBeDefined()
    await (wrapper.vm as any).startTest()
    expect(global.fetch).not.toHaveBeenCalled()

    // 即使在 watcher 更新之前仍持有旧图片值，提交入口也必须拦截。
    ;(wrapper.vm as any).selectedModelId = 'gpt-image-2'
    await (wrapper.vm as any).startTest()
    expect(global.fetch).not.toHaveBeenCalled()
  })

  it('模型异步加载完成时会保留糖果模式的文本模型限制', async () => {
    let resolveModels!: (models: Array<{ id: string; display_name: string }>) => void
    getAvailableModels.mockReturnValue(new Promise((resolve) => { resolveModels = resolve }))
    const wrapper = mountModal({
      id: 46, name: 'OpenAI OAuth', platform: 'openai', type: 'oauth', status: 'active'
    })
    await wrapper.setProps({ show: true })
    await wrapper.findAll('select.select-stub')[1].setValue('candy')
    resolveModels([
      { id: 'gpt-image-2', display_name: 'GPT Image 2' },
      { id: 'gpt-5.4', display_name: 'GPT-5.4' }
    ])
    await flushPromises()

    expect((wrapper.vm as any).selectedModelId).toBe('gpt-5.4')
    expect(wrapper.findAll('select.select-stub')[0].findAll('option').map((option) => option.attributes('value'))).toEqual(['gpt-5.4'])
  })

  it('renders Chat Completions path status from test SSE', async () => {
    getAvailableModels.mockResolvedValue([{ id: 'gpt-5.4', display_name: 'GPT-5.4' }])
    global.fetch = vi.fn().mockResolvedValue(createStreamResponse([
      'data: {"type":"status","text":"已通过 /v1/chat/completions 验证"}\n\n',
      'data: {"type":"test_complete","success":true}\n\n'
    ])) as any
    const wrapper = mountModal({
      id: 47, name: 'OpenAI API Key', platform: 'openai', type: 'apikey', status: 'active'
    })
    await wrapper.setProps({ show: true })
    await flushPromises()
    await (wrapper.vm as any).startTest()
    await flushPromises()

    expect(wrapper.text()).toContain('已通过 /v1/chat/completions 验证')
  })
})
