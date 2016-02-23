<div class="container" ng-controller="VoteCtrl">
  <a href="http://sg.uchicago.edu">
    <img src="/sg.png" style="width:3in" class="masthead hidden-xs hidden-sm"/>
  </a>
  <div class="alert alert-info" ng-show="vote.join('') && !pending">
	You are confirmed for
    <strong ng-repeat="cand in vote track by $index">{{
        $index > 0 ? "," : "" }}
      {{ getCandidateFor(cand).name }}</strong>.
	<a ng-click="voteFor(null)" class="btn btn-default">Cancel Selection</a>
  </div>
  <div class="panel panel-default" ng-show="hasDescription && !editMode">
    <div class="panel-body" ng-bind-html="markdown(descriptionText)"></div>
  </div>
  <div class="panel panel-default">
    <div class="panel-heading" ng-show="isAdmin">
      <h3 class="panel-title">
      <a class="btn btn-default"
         href="{{ baseURL + '/results.csv' }}">Download CSV</a>
      <a class="btn btn-default vote-edit-button"
         ng-click="setEditMode(!editMode);">
           {{ editMode ? "Save" : "Edit" }}</a>
      </h3>
    </div>
    <table class="table">
      <tr ng-show="editMode">
        <td>
          <div class="btn-group" dropdown>
            <button class="btn btn-default dropdown-toggle" dropdown-toggle>
              Add Plugin...
              <span class="caret"></span></button>
            <ul class="dropdown-menu">
              <li class="{{ voteLimits ? 'disabled' : '' }}">
                <a ng-click="setOption('limits', true)">Vote Limits</a></li>
              <li class="{{ voteWeight > 0 ? 'disabled' : '' }}">
                <a ng-click="setOption('weight', true)">
                   Dollar Amount Per Vote</a></li>
              <li class="{{ voteRandomized ? 'disabled' : '' }}">
                <a ng-click="setOption('randomized', true)">
                  Randomized Voting</a></li>
              <li class="{{ showProgress ? 'disabled' : '' }}">
                <a ng-click="setOption('showProgress', true)">
                  Progress Bar</a></li>
              <li class="{{ hasDescription ? 'disabled' : '' }}">
                <a ng-click="setOption('hasDescription', true)">
                  Description</a></li>
              <li class="{{ voteLimit > 1 ? 'disabled' : '' }}">
                <a ng-click="setOption('voteLimit', true)">
                  Rank Several</a></li>
            </ul>
          </div>
        </td>
      </tr>
      <tr ng-show="editMode && voteLimits" class="vote-plugin">
        <td>
          <label>
            Users cannot select above the limit.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('limits', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && voteWeight != 0" class="form-inline vote-plugin">
        <td>
          <label>
            Each selection contributes:
          </label>
        </td>
        <td>
          <input type="number" class="form-control" style="width:5em"
            ng-model="voteWeight"/>
        </td>
      </tr>
      <tr ng-show="editMode && voteRandomized" class="form-inline vote-plugin">
        <td>
          <label>
            Options are randomized before display.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('randomized', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && showProgress" class="form-inline vote-plugin">
        <td>
          <label>
            Voters can see the number of spaces left.
          </label>
        </td>
        <td>
          <a class="btn btn-danger" ng-click="setOption('showProgress', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && hasDescription" class="vote-plugin">
        <td>
          <textarea placeholder="(enter a description)"
             ng-model="descriptionText"
             class="form-control"></textarea>
        </td>
        <td>
          <a class="btn btn-danger"
             ng-click="setOption('hasDescription', false)">
            Disable
          </a>
        </td>
      </tr>
      <tr ng-show="editMode && voteLimit > 1" class="form-inline vote-plugin">
        <td>
          <label>
            Voters can rank this many options:
          </label>
        </td>
        <td>
          <input type="number" class="form-control" style="width:5em"
            ng-model="voteLimit"/>
        </td>
      </tr>
      <tbody ng-repeat="(key, section) in candidates | groupBy:(editMode ? '' : 'section')">
        <tr class="{{ expanded[key] ? 'expanded' : 'collapsed' }}"
            ng-show="!editMode && key"
            ng-click="expand(key)"><th colspan="2">
              <span>{{ expanded[key] ? "\u25BC" : "\u25B6" }}</span>
              {{ key }}</th></tr>
        <tr ng-show="expanded[key] || !key || editMode"
            ng-repeat="candidate in section |
                orderBy:(!editMode && voteRandomized ? ['section','order'] : '')">
          <td ng-hide="editMode">
            <p><strong>{{ candidate.name }}</strong>
              <span ng-bind-html="markdown(candidate.description)">
                {{ candidate.description }}</span></p>
          </td>
          <td ng-show="editMode">
            <input type="text" class="form-control"
                   ng-model="candidate.name"
                   placeholder="(name)"/>
            <textarea class="form-control"
                      ng-model="candidate.description"
                      placeholder="(description)"></textarea>
          </td>
          <td style="text-align:right" ng-hide="editMode" class="vote-right">
            <div class="btn-group vote-button" ng-show="vote.length > 1">
              <a ng-repeat="cand in vote track by $index"
                 class="btn {{ cand == candidate.id ?
                            (pending ? 'btn-success disabled' :
                              'btn-success active')
                              : ((candidate.progress + 1) * (voteWeight || 1)
                                  > candidate.request ?
                                  'btn-default disabled'
                                  : 'btn-default') }}"
                 ng-click="voteFor(candidate, $index)">#{{$index + 1}}</a>
            </div>
            <a class="btn vote-button
                        {{ vote == candidate.id ?
                            (pending ? 'btn-primary disabled' :
                              'btn-success active')
                              : ((candidate.progress + 1) * (voteWeight || 1)
                                  > candidate.request ?
                                  'btn-primary disabled'
                                  : 'btn-primary')
                          }}"
                 ng-click="voteFor(candidate, 0)"
                 ng-show="vote.length == 1">
                {{ getVerbFor(candidate) }}
            </a>
            <div class="progress vote-progress" ng-show="candidate.request && (showProgress || isAdmin)">
              <div class="progress-bar progress-bar-success progress-bar-striped"
                   style="
                     width:{{ candidate.progress * (voteWeight || 1) * 100
                         / candidate.request }}%;">
              </div>
            </div>
            <div class="vote-caption" ng-show="candidate.request && (showProgress || isAdmin)">
              {{ candidate.progress * (voteWeight || 1) | number:0 }} of
              {{ candidate.request | number:0 }}
            </div>
          </td>
          <td ng-show="editMode" class="vote-right">
            <input type="number" class="form-control"
                   ng-model="candidate.request" placeholder="(no target)"/>
            <input type="text" class="form-control"
                   ng-model="candidate.section" placeholder="(no section)"/>
          </td>
        </tr>
      </tbody>
      <tr ng-show="editMode">
        <td colspan="2">
            <a class="btn btn-primary"
              ng-click="addNewRow()">Add New Row</a></td>
      </tr>
    </table>
  </div>
</div>
