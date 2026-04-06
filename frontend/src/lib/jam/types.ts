export interface Claims {
  userId: string;
  plan: string;
  sessionState: string;
}

export interface SessionParticipant {
  userId: string;
  role: string;
  muted?: boolean;
}

export interface SessionSnapshot {
  jamId: string;
  status: string;
  hostUserId: string;
  participants: SessionParticipant[];
  sessionVersion: number;
  endCause?: string;
  endedBy?: string;
}

export interface QueueItem {
  itemId: string;
  trackId: string;
  addedBy: string;
}

export interface QueueSnapshot {
  jamId: string;
  queueVersion: number;
  items: QueueItem[];
}

export interface SessionStateSnapshot {
  session: SessionSnapshot;
  queue: QueueSnapshot;
  aggregateVersion: number;
}

export interface TrackLookup {
  trackId: string;
  isPlayable: boolean;
  reasonCode?: string;
  title?: string;
  artist?: string;
}

export interface PlaybackCommandRequest {
  command: string;
  trackId?: string;
  clientEventId: string;
  expectedQueueVersion: number;
  positionMs?: number;
}

export interface PlaybackAccepted {
  accepted: boolean;
}

export interface BffIssue {
  dependency: string;
  code: string;
  message: string;
}

export interface BffOrchestrationData {
  claims: Claims;
  sessionState: SessionStateSnapshot;
  track?: TrackLookup;
  partial: boolean;
  dependencyStatuses: Record<string, string>;
  issues?: BffIssue[];
}
