import api from './api';

export default {
  // Get long-term memory content
  getLongTermMemory() {
    return api.get('/memory/longterm');
  },

  // Update long-term memory content
  updateLongTermMemory(content) {
    return api.post('/memory/longterm', { content });
  },

  // Get daily notes (default 7 days)
  getDailyNotes(days = 7) {
    return api.get(`/memory/daily?days=${days}`);
  }
};
