import { ref } from 'tsx-vanilla';

import { ActionId } from '../../proto_utils/action_id.js';
import { TypedEvent } from '../../typed_event.js';
import { isRightClick } from '../../utils.js';
import { Input, InputConfig } from '../input.js';

// Data for creating an icon-based input component.
//
// E.g. one of these for arcane brilliance, another for kings, etc.
// ModObject is the object being modified (Sim, Player, or Target).
// ValueType is either number or boolean.
export interface IconPickerConfig<ModObject, ValueType> extends InputConfig<ModObject, ValueType> {
	actionId: ActionId;

	// The number of possible 'states' this icon can have. Most inputs will use 2
	// for a bi-state icon (on or off). 0 indicates an unlimited number of states.
	states: number;

	// Only used if states >= 3.
	improvedId?: ActionId;

	// Only used if states >= 4.
	improvedId2?: ActionId;
}

// Icon-based UI for picking buffs / consumes / etc
// ModObject is the object being modified (Sim, Player, or Target).
export class IconPicker<ModObject, ValueType> extends Input<ModObject, ValueType> {
	active: boolean;

	private readonly config: IconPickerConfig<ModObject, ValueType>;

	private readonly rootAnchor: HTMLAnchorElement;
	private readonly improvedAnchor: HTMLAnchorElement;
	private readonly improvedAnchor2: HTMLAnchorElement;
	private readonly counterElem: HTMLElement;

	private currentValue: number;
	private storedValue: ValueType | undefined;

	constructor(parent: HTMLElement, modObj: ModObject, config: IconPickerConfig<ModObject, ValueType>) {
		super(parent, 'icon-picker-root', modObj, config);
		this.rootElem.classList.add('icon-picker');
		this.active = false;
		this.config = config;
		this.currentValue = 0;

		this.rootAnchor = document.createElement('a');
		this.rootAnchor.classList.add('icon-picker-button');
		this.rootAnchor.dataset.whtticon = 'false';
		this.rootAnchor.dataset.disableWowheadTouchTooltip = 'true';
		this.rootAnchor.target = '_blank';
		this.rootElem.prepend(this.rootAnchor);

		const useImprovedIcons = Boolean(this.config.improvedId);
		if (useImprovedIcons) {
			this.rootAnchor.classList.add('use-improved-icons');
		}
		if (this.config.improvedId2) {
			this.rootAnchor.classList.add('use-improved-icons2');
		}
		if (!useImprovedIcons && this.config.states > 2) {
			this.rootAnchor.classList.add('use-counter');
		}

		const ia = ref<HTMLAnchorElement>();
		const ia2 = ref<HTMLAnchorElement>();
		const ce = ref<HTMLSpanElement>();
		this.rootAnchor.appendChild(
			<div className="icon-input-level-container">
				<a
					ref={ia}
					className="icon-picker-button icon-input-improved icon-input-improved1"
					dataset={{ whtticon: 'false', disableWowheadTouchTooltip: 'true' }}></a>
				<a
					ref={ia2}
					className="icon-picker-button icon-input-improved icon-input-improved2"
					dataset={{ whtticon: 'false', disableWowheadTouchTooltip: 'true' }}></a>
				<span ref={ce} className={`icon-picker-label ${this.config.states > 2 ? '' : 'hide'}`}></span>
			</div>,
		);

		this.improvedAnchor = ia.value!;
		this.improvedAnchor2 = ia2.value!;
		this.counterElem = ce.value!;

		this.updateButtonImage()

		const event = this.config.changedEvent(this.modObject).on(() => {
			this.updateButtonImage()

			if (this.showWhen()) {
				this.rootElem.classList.remove('hide');
				this.restoreValue();
			} else {
				this.storeValue();
				this.rootElem.classList.add('hide');
			}
		});

		this.init();

		this.rootAnchor.addEventListener(
			'click',
			event => {
				event.preventDefault();
				this.handleLeftClick();
			},
			{ signal: this.signal },
		);

		this.rootAnchor.addEventListener(
			'contextmenu',
			event => {
				event.preventDefault();
			},
			{ signal: this.signal },
		);

		this.rootAnchor.addEventListener(
			'mousedown',
			event => {
				const rightClick = isRightClick(event);

				if (rightClick) {
					event.preventDefault();
					this.handleRightClick();
				}
			},
			{ signal: this.signal },
		);

		this.addOnDisposeCallback(() => event.dispose());
	}

	updateButtonImage() {
		this.config.actionId.fillAndSet(this.rootAnchor, true, true);

		if (this.config.states >= 3 && this.config.improvedId) {
			this.config.improvedId.fillAndSet(this.improvedAnchor, true, true, { signal: this.signal });
		}
		if (this.config.states >= 4 && this.config.improvedId2) {
			this.config.improvedId2.fillAndSet(this.improvedAnchor2, true, true, { signal: this.signal });
		}
	}

	handleLeftClick() {
		if (this.config.states == 0 || this.currentValue + 1 < this.config.states) {
			this.currentValue++;
			this.inputChanged(TypedEvent.nextEventID());
		} else if (this.currentValue > 0) {
			// roll over
			this.currentValue = 0;
			this.inputChanged(TypedEvent.nextEventID());
		}
	}

	handleRightClick() {
		if (this.currentValue > 0) {
			this.currentValue--;
		} else {
			// roll over
			if (this.config.states === 0) {
				this.currentValue = 1;
			} else {
				this.currentValue = this.config.states - 1;
			}
		}
		this.inputChanged(TypedEvent.nextEventID());
	}

	getInputElem(): HTMLElement {
		return this.rootAnchor;
	}

	getInputValue(): ValueType {
		if (this.config.states == 2) {
			return Boolean(this.currentValue) as unknown as ValueType;
		} else {
			return this.currentValue as unknown as ValueType;
		}
	}

	// Returns the ActionId of the currently selected value, or null if none selected.
	getActionId(): ActionId | null {
		// Go directly to source because during event propogation
		//  the internal `this.currentValue` may not yet have been updated.
		const v = Number(this.config.getValue(this.modObject));
		if (v == 0) {
			return null;
		} else if (v == 1) {
			return this.config.actionId;
		} else if (v == 2 && this.config.improvedId) {
			return this.config.improvedId;
		} else if (v == 3 && this.config.improvedId2) {
			return this.config.improvedId2;
		} else {
			return this.config.actionId;
		}
	}

	setInputValue(newValue: ValueType) {
		this.currentValue = Number(newValue);

		if (this.currentValue > 0) {
			this.active = true;
			this.rootAnchor.classList.add('active');
			this.counterElem.classList.add('active');
		} else {
			this.active = false;
			this.rootAnchor.classList.remove('active');
			this.counterElem.classList.remove('active');
		}

		if (this.config.states >= 3 && this.config.improvedId) {
			if (this.currentValue > 1) {
				this.improvedAnchor.classList.add('active');
			} else {
				this.improvedAnchor.classList.remove('active');
			}
		}
		if (this.config.states >= 4 && this.config.improvedId2) {
			if (this.currentValue > 2) {
				this.improvedAnchor2.classList.add('active');
				this.improvedAnchor.hidden = true;
				this.improvedAnchor2.hidden = false;
			} else {
				this.improvedAnchor2.classList.remove('active');
				this.improvedAnchor.hidden = false;
				this.improvedAnchor2.hidden = true;
			}
		}

		if (!this.config.improvedId && (this.config.states > 3 || this.config.states == 0)) {
			this.counterElem.textContent = String(this.currentValue);
		}
	}

	/**
	 * Stores value of current input and hides the element for later
	 * restoration. Useful for events which trigger the element
	 * on and off.
	 */
	storeValue() {
		if (typeof this.storedValue !== 'undefined') return;

		this.storedValue = this.getInputValue();
		this.setInputValue(0 as ValueType);
		this.inputChanged(TypedEvent.nextEventID());
	}

	/**
	 * Restores value of current input and shows the element.
	 */
	restoreValue() {
		if (typeof this.storedValue === 'undefined') return;

		this.setInputValue(this.storedValue);
		this.inputChanged(TypedEvent.nextEventID());
		this.storedValue = undefined;
	}

	showWhen() {
		return this.config.actionId != null && (!this.config.showWhen || this.config.showWhen(this.modObject));
	}
}
