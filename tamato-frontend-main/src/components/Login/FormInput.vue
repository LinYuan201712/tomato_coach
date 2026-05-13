<template>
  <div class="form-input-group">
    <label v-if="label" :for="id" class="input-label">
      {{ label }}
      <span v-if="required" class="required">*</span>
    </label>
    
    <div class="input-container" :class="{ 'error': error, 'focused': isFocused }">
      <input
        :id="id"
        :type="internalType"
        :value="modelValue"
        :placeholder="placeholder"
        :required="required"
        :disabled="disabled"
        @input="$emit('update:modelValue', $event.target.value)"
        @focus="isFocused = true"
        @blur="isFocused = false"
        class="form-input"
      />
      
      <div v-if="showPasswordToggle && type === 'password'" class="password-toggle">
        <button
          type="button"
          @click="togglePasswordVisibility"
          class="toggle-btn"
        >
          {{ internalType === 'password' ? '显示' : '隐藏' }}
        </button>
      </div>
    </div>
    
    <div v-if="error" class="error-message">
      {{ error }}
    </div>
    
    <div v-if="hint && !error" class="hint-message">
      {{ hint }}
    </div>
  </div>
</template>

<script>
export default {
  name: 'FormInput',
  props: {
    modelValue: {
      type: [String, Number],
      default: ''
    },
    type: {
      type: String,
      default: 'text',
      validator: (value) => ['text', 'password', 'email', 'number', 'tel'].includes(value)
    },
    label: {
      type: String,
      default: ''
    },
    placeholder: {
      type: String,
      default: ''
    },
    required: {
      type: Boolean,
      default: false
    },
    disabled: {
      type: Boolean,
      default: false
    },
    error: {
      type: String,
      default: ''
    },
    hint: {
      type: String,
      default: ''
    },
    showPasswordToggle: {
      type: Boolean,
      default: true
    },
    id: {
      type: String,
      default: () => `input-${Math.random().toString(36).substr(2, 9)}`
    }
  },
  emits: ['update:modelValue'],
  data() {
    return {
      isFocused: false,
      internalType: this.type
    }
  },
  methods: {
    togglePasswordVisibility() {
      this.internalType = this.internalType === 'password' ? 'text' : 'password'
    }
  },
  watch: {
    type(newVal) {
      this.internalType = newVal
    }
  }
}
</script>

<style scoped>
.form-input-group {
  margin-bottom: 1.5rem;
}

.input-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #333;
  font-size: 0.9rem;
}

.required {
  color: #e74c3c;
}

.input-container {
  position: relative;
  display: flex;
  align-items: center;
  border: 2px solid #e9ecef;
  border-radius: 6px;
  transition: all 0.3s ease;
  background: white;
}

.input-container.focused {
  border-color: #eeaa67;
  box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
}

.input-container.error {
  border-color: #e74c3c;
}

.form-input {
  flex: 1;
  padding: 12px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 1rem;
  width: 100%;
}

.form-input::placeholder {
  color: #adb5bd;
}

.form-input:disabled {
  background-color: #f8f9fa;
  color: #6c757d;
  cursor: not-allowed;
}

.password-toggle {
  padding-right: 12px;
}

.toggle-btn {
  background: none;
  border: none;
  color: #cc2a1f;
  cursor: pointer;
  font-size: 0.8rem;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background 0.2s ease;
}

.toggle-btn:hover {
  background: #f8f9fa;
}

.error-message {
  margin-top: 0.5rem;
  color: #e74c3c;
  font-size: 0.8rem;
}

.hint-message {
  margin-top: 0.5rem;
  color: #6c757d;
  font-size: 0.8rem;
}
</style>